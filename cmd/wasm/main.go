//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"
)

type Profile struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
	Type    string   `json:"content_type"`
}

func probe(this js.Value, args []js.Value) interface{} {
	url := args[0].String()

	p := Profile{
		URL:     url,
		Methods: []string{"GET", "POST"},
		Type:    "application/json",
	}

	b, _ := json.MarshalIndent(p, "", "  ")
	return string(b)
}

func main() {
	js.Global().Set("restlessProbe", js.FuncOf(probe))
	select {}
}
