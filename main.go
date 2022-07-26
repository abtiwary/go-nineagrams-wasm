package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"syscall/js"
	"time"
)

//go:embed nineletterwords.json
var fs embed.FS

type WordInfo struct {
	Word  string `json:"word"`
	Rank  string `json:"rank"`
	Count string `json:"count"`
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
	solutions := make(map[string]struct{})
	solutionList := make([]interface{}, 0)
	for _, s := range solutionInfo {
		solutions[s.Word] = struct{}{}
		solutionList = append(solutionList, ToBase64(s.Word))
	}

	var puzzleWord string
	for {
		// shuffle the randomKey, then check that it isn't one of the solutions
		shuffledRandomKey := ShuffleKey(randomKey)

		if _, ok := solutions[shuffledRandomKey]; !ok {
			puzzleWord = shuffledRandomKey
			break
		}
	}
	return puzzleWord, solutionList
}

func ComputeAPuzzleWord(this js.Value, args []js.Value) interface{} {
	data, _ := fs.ReadFile("nineletterwords.json")

	// read the embedded JSON document
	var words map[string][]WordInfo
	_ = json.Unmarshal(data, &words)

	// extract a list of all the keys
	// each key is nine letter word - but sorted
	keys := make([]string, 0, len(words))
	for k := range words {
		keys = append(keys, k)
	}

	// next, pick a key at random
	randomKey := GetRandomKey(keys)
	solutions := words[randomKey]
	// get the final word to use in the puzzle
	puzzleWord, solutionList := GetPuzzleWord(randomKey, solutions)

	js.Global().Set("puzzle_word", puzzleWord)
	js.Global().Set("puzzle_key", randomKey)

	return solutionList
}

func main() {
	c := make(chan bool)
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
