package main

import (
	"compress/gzip"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"syscall/js"
	"time"
)

//go:embed *.json.gz
var fs embed.FS

var (
	difficulties map[int]string
	words        map[int]map[string][]WordInfo
	puzzleKeys   map[int][]string
	solutionList []interface{}
	puzzleWord   string
	puzzleKey    string
)

type WordInfo struct {
	Word    string `json:"word"`
	Count   string `json:"count"`
	Percent string `json:"percent"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func PrintWASMLoadStatus(this js.Value, args []js.Value) interface{} {
	return "WASM loaded!"
}

func GetRandomKey(keys []string) string {
	randomNum := rand.Intn(len(keys)-0) + 0
	return keys[randomNum]
}

func ShuffleKey(key string) string {
	keyRunes := []rune(key)
	rand.Shuffle(len(keyRunes), func(i, j int) {
		keyRunes[i], keyRunes[j] = keyRunes[j], keyRunes[i]
	})
	return string(keyRunes)
}

func ToBase64(toEncode string) string {
	encodedStr := base64.StdEncoding.EncodeToString([]byte(toEncode))
	return encodedStr
}

func FromBase64(toDecode string) string {
	decodedStr, _ := base64.StdEncoding.DecodeString(toDecode)
	return string(decodedStr)
}

func ToBase64JS(this js.Value, args []js.Value) interface{} {
	toEncode := args[0].Get("to_encode").String()
	return ToBase64(toEncode)
}

func FromBase64JS(this js.Value, args []js.Value) interface{} {
	toDecode := args[0].Get("to_decode").String()
	return FromBase64(toDecode)
}

func GetPuzzleWord(randomKey string, solutionInfo []WordInfo) (string, []interface{}) {
	solutionSet := make(map[string]struct{})
	listOfSolutions := make([]interface{}, 0)
	for _, s := range solutionInfo {
		solutionSet[s.Word] = struct{}{}
		listOfSolutions = append(listOfSolutions, ToBase64(s.Word))
	}

	var word string
	for {
		// shuffle the randomKey, then check that it isn't one of the solutions
		shuffledRandomKey := ShuffleKey(randomKey)

		if _, ok := solutionSet[shuffledRandomKey]; !ok {
			word = shuffledRandomKey
			break
		}
	}
	return word, listOfSolutions
}

func ComputeAPuzzleWord(this js.Value, args []js.Value) interface{} {
	// get the game difficulty
	gameDifficulty := js.Global().Get("game_difficulty").Int()
	fmt.Println(gameDifficulty)

	// pick a key at random
	puzzleKey = GetRandomKey(puzzleKeys[gameDifficulty])
	solutions := words[gameDifficulty][puzzleKey]

	// get the final word to use in the puzzle
	puzzleWord, solutionList = GetPuzzleWord(puzzleKey, solutions)

	js.Global().Set("puzzle_word", puzzleWord)
	js.Global().Set("puzzle_key", puzzleKey)

	wordsAsJson := make(map[string][]string)
	for k := range words[gameDifficulty] {
		wordsAsJson[k] = make([]string, 0)
		for _, v := range words[gameDifficulty][k] {
			wordsAsJson[k] = append(wordsAsJson[k], ToBase64(v.Word))
		}
	}
	wordsAsJsonStr, _ := json.Marshal(wordsAsJson)
	js.Global().Set("puzzle_data_str", string(wordsAsJsonStr))

	return solutionList
}

func InitializeApp() {
	for k, v := range difficulties {
		jsonFile, _ := fs.Open(v)
		gz, _ := gzip.NewReader(jsonFile)
		jsonBody, _ := io.ReadAll(gz)
		wordsAtK := make(map[string][]WordInfo)
		_ = json.Unmarshal(jsonBody, &wordsAtK)
		words[k] = wordsAtK

		// extract a list of all the keys
		// each key is an X letter word - but sorted
		pKeys := make([]string, 0)
		for k := range words[k] {
			pKeys = append(pKeys, k)
		}
		puzzleKeys[k] = pKeys
	}
}

func main() {
	c := make(chan bool)

	difficulties = make(map[int]string)
	difficulties[1] = "four_letters.json.gz"
	difficulties[2] = "five_letters.json.gz"
	difficulties[3] = "eight_letters.json.gz"
	difficulties[4] = "nine_letters.json.gz"

	words = make(map[int]map[string][]WordInfo)
	puzzleKeys = make(map[int][]string)
	solutionList = make([]interface{}, 0)

	InitializeApp()

	js.Global().Set("PrintWASMLoadStatus", js.FuncOf(PrintWASMLoadStatus))
	js.Global().Set("ComputeAPuzzleWord", js.FuncOf(ComputeAPuzzleWord))
	js.Global().Set("ToBase64JS", js.FuncOf(ToBase64JS))
	js.Global().Set("FromBase64JS", js.FuncOf(FromBase64JS))

	<-c
}

//
// Compile with:
//    GOOS=js GOARCH=wasm go build -o main.wasm
// last tested on Go 1.18
//
