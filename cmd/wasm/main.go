//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"github.com/Mr-Fullstack/orchaxon-autotest/pkg/core"
)

func main() {
	// Canal para manter o programa rodando
	c := make(chan struct{}, 0)
	
	// Registra a função no JavaScript global
	js.Global().Set("GenerateTestCode", js.FuncOf(GenerateWrapper))
	
	fmt.Println("✅ AutoTest Gen WASM Initialized")
	<-c
}

// Essa função recebe o JSON do JavaScript e devolve o Código Gerado
func GenerateWrapper(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "Error: No input provided"
	}
	jsonInput := args[0].String()
	
	var config core.MetaFramework 
	err := json.Unmarshal([]byte(jsonInput), &config)
	if err != nil {
		return fmt.Sprintf("// Error parsing JSON: %s", err.Error())
	}

	// Verifica linguagens
	languages := config.Meta.Langs
	if len(languages) == 0 {
		if config.Meta.Lang != "" {
			languages = []string{config.Meta.Lang}
		} else {
			return "// Error: No language specified in 'meta.lang' or 'meta.langs'"
		}
	}

	// Gera o código para todas as linguagens pedidas e concatena
	var finalOutput string
	for _, lang := range languages {
		code, err := core.ProcessTemplate(config, lang)
		if err != nil {
			finalOutput += fmt.Sprintf("// Error generating %s: %s\n\n", lang, err.Error())
		} else {
			finalOutput += fmt.Sprintf("// --- Generated [%s] ---\n%s\n\n", lang, code)
		}
	}

	return finalOutput
}