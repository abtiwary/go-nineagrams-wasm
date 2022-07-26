package main

import (
	"embed"
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

func GetPuzzleWord(randomKey string, solutionInfo []WordInfo) (string, []interface{}) {
	solutions := make(map[string]struct{})
	solutionList := make([]interface{}, 0)
	for _, s := range solutionInfo {
		solutions[s.Word] = struct{}{}
		solutionList = append(solutionList, s.Word)
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
	<-c
}

//
// Compile with:
//    GOOS=js GOARCH=wasm go build -o main.wasm
// last tested on Go 1.18
//
