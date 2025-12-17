package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	// IMPORTANTE: Importando seu pacote core
	// Certifique-se que o go.mod tem esse nome de módulo
	"github.com/Mr-Fullstack/orchaxon-autotest/pkg/core"
)

// Helper para determinar nome do arquivo (Só usado na CLI)
var extensions = map[string]string{
	"go": "go", "csharp": "cs", "cs": "cs", "kotlin": "kt", "java": "java",
	"php": "php", "typescript": "ts", "ts": "ts", "node": "test.js", "js": "test.js",
    "python": "py", "py": "py",
}

func getFilename(className, lang string) string {
	ext, exists := extensions[lang] // removido strings.ToLower para simplificar, ajuste se precisar
	if !exists { ext = "txt" }
	
    suffix := "Test"
	if ext == "test.js" { suffix = "" }
	if ext == "go" { suffix = "_test" }
    if ext == "py" { return fmt.Sprintf("test_%s.py", className) } // Exemplo Python
	
	return fmt.Sprintf("%s%s.%s", className, suffix, ext)
}

func main() {
	start := time.Now()

	// Flags
	fileFlag := flag.String("file", "", "Path to JSON spec file")
	langFlag := flag.String("lang", "", "Language")
	classFlag := flag.String("class", "", "Class Name")
	printFlag := flag.Bool("print", false, "Print to console")
	flag.Parse()

    // Configura pasta de saída
	const outputDir = "test"
	if !*printFlag {
		os.MkdirAll(outputDir, 0755)
	}

	// --- MODO 1: ARQUIVO JSON ---
	if *fileFlag != "" {
		data, err := os.ReadFile(*fileFlag)
		if err != nil {
			fmt.Printf("❌ Error reading file: %v\n", err)
			os.Exit(1)
		}

        // CORREÇÃO: Usando core.MetaFramework
		var config core.MetaFramework 
		if err := json.Unmarshal(data, &config); err != nil {
			fmt.Printf("❌ Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		langs := config.Meta.Langs
		if len(langs) == 0 && config.Meta.Lang != "" {
			langs = []string{config.Meta.Lang}
		}

		for _, lang := range langs {
            // CORREÇÃO: Chamando core.ProcessTemplate (com P maiúsculo)
			code, err := core.ProcessTemplate(config, lang)
			if err != nil {
				fmt.Printf("⚠️ Skipping %s: %v\n", lang, err)
				continue
			}
            
            filename := getFilename(config.Target.ClassName, lang)
            finalPath := filepath.Join(outputDir, filename)
            os.WriteFile(finalPath, []byte(code), 0644)
            fmt.Printf("✓ Generated %s\n", finalPath)
		}
		return
	}

	// --- MODO 2: SIMPLE CLI ---
	if *langFlag != "" && *classFlag != "" {
        // CORREÇÃO: Usando as structs do pacote core
		fakeConfig := core.MetaFramework{
			Meta: core.MetaInfo{Lang: *langFlag},
			Target: core.TargetInfo{ClassName: *classFlag, MethodName: "MyMethod"},
			Scenarios: []core.Scenario{{ID: "should_work", Description: "Should return expected result"}},
		}

        // CORREÇÃO: Chamando core.ProcessTemplate
		code, err := core.ProcessTemplate(fakeConfig, *langFlag)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		if *printFlag {
			fmt.Println(code)
		} else {
            filename := getFilename(*classFlag, *langFlag)
            finalPath := filepath.Join(outputDir, filename)
            os.WriteFile(finalPath, []byte(code), 0644)
			fmt.Printf("✓ Generated %s (%.2fs)\n", finalPath, time.Since(start).Seconds())
		}
		return
	}

	fmt.Println("Usage: autotest -file specs.json OR -lang go -class User")
}