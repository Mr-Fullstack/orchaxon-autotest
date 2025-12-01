package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// --- 1. ESTRUTURAS DE DADOS (Meta-Framework) ---

type MetaFramework struct {
	Meta         MetaInfo     `json:"meta"`
	Target       TargetInfo   `json:"target"`
	Dependencies []Dependency `json:"dependencies"`
	Scenarios    []Scenario   `json:"scenarios"`
}

type MetaInfo struct {
	Lang  string   `json:"lang"`  // Legado (Single)
	Langs []string `json:"langs"` // Novo (Multi-language support)
}

type TargetInfo struct {
	ClassName  string `json:"class_name"`
	MethodName string `json:"method_name"`
}

type Dependency struct {
	FieldName     string `json:"field_name"`
	InterfaceName string `json:"interface_name"`
}

type Scenario struct {
	ID           string        `json:"id"`
	Description  string        `json:"description"`
	MocksSetup   []MockSetup   `json:"mocks_setup"`
	Expectations Expectation   `json:"expectations"`
}

type MockSetup struct {
	Dependency  string      `json:"dependency"`
	Method      string      `json:"method"`
	ReturnValue interface{} `json:"return_value"`
}

type Expectation struct {
	ReturnValue interface{} `json:"return_value"`
}

// --- 2. TEMPLATES MULTI-LINGUAGEM ---

// TEMPLATE GO
const goTmpl = `package {{.Target.ClassName | ToLower}}

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

{{- if .Dependencies}}
// Mocks Definitions
{{- range .Dependencies}}
type Mock{{.InterfaceName}} struct {
	mock.Mock
}
{{- end}}
{{- end}}

func Test{{.Target.MethodName}}(t *testing.T) {
	{{- if not .Scenarios}}
	// Simple test case
	t.Run("should work correctly", func(t *testing.T) {
		// Arrange
		// sut := New{{.Target.ClassName}}()

		// Act
		// result := sut.{{.Target.MethodName}}()

		// Assert
		// assert.NotNil(t, result)
	})
	{{- end}}

	{{- range $s := .Scenarios}}
	t.Run("{{$s.Description}}", func(t *testing.T) {
		// Arrange
		{{- range $dep := $.Dependencies}}
		// mock{{$dep.FieldName}} := new(Mock{{$dep.InterfaceName}})
		{{- end}}

		{{- range $m := $s.MocksSetup}}
		// {{.Dependency}}.On("{{.Method}}").Return({{.ReturnValue | FormatValue}})
		{{- end}}

		// Act
		// result := sut.{{$.Target.MethodName}}()

		// Assert
		{{- if $s.Expectations.ReturnValue}}
		// assert.Equal(t, {{$s.Expectations.ReturnValue | FormatValue}}, result)
		{{- end}}
	})
	{{- end}}
}`

const csharpTmpl = `using Xunit;
using NSubstitute;

namespace Tests
{
    public class {{.Target.ClassName}}Tests
    {
        {{- range .Dependencies}}
        private readonly {{.InterfaceName}} _{{.FieldName}};
        {{- end}}
        // private readonly {{.Target.ClassName}} _sut;
    
        public {{.Target.ClassName}}Tests()
        {
            {{- range .Dependencies}}
            _{{.FieldName}} = Substitute.For<{{.InterfaceName}}>();
            {{- end}}
            // _sut = new {{.Target.ClassName}}({{range $i, $e := .Dependencies}}{{if $i}}, {{end}}_{{$e.FieldName}}{{end}});
        }
    
        {{- if not .Scenarios}}
        [Fact]
        public void Should_DoWork()
        {
            // Arrange
            // Act
            // var result = _sut.{{.Target.MethodName}}();
            // Assert
        }
        {{- end}}

        {{- range .Scenarios}}
        [Fact(DisplayName = "{{.Description}}")]
        public void {{.ID | ToPascal}}()
        {
            // Arrange
            {{- range .MocksSetup}}
            _{{.Dependency}}.{{.Method}}(Arg.Any<object>()).Returns({{.ReturnValue | FormatValue}});
            {{- end}}
    
            // Act
            // var result = _sut.{{$.Target.MethodName}}();
    
            // Assert
            {{- if .Expectations.ReturnValue}}
            // Assert.Equal({{.Expectations.ReturnValue | FormatValue}}, result);
            {{- end}}
        }
        {{- end}}
    }
}`

const nodeNativeTmpl = `import { describe, it, mock } from 'node:test';
import assert from 'node:assert';
// import { {{.Target.ClassName}} } from '../src/{{.Target.ClassName}}.js'; 

describe('{{.Target.ClassName}}', () => {
    
    {{- if not .Scenarios}}
    it('should execute correctly', () => {
        // const sut = new {{.Target.ClassName}}();
        // const result = sut.{{.Target.MethodName}}();
        // assert.ok(result);
    });
    {{- end}}

    {{- range $s := .Scenarios}}
    it('{{$s.Description}}', () => {
        // Arrange
        {{- range $dep := $.Dependencies}}
        const {{$dep.FieldName}} = {
             {{- range $m := $s.MocksSetup}}
                {{- if eq $m.Dependency $dep.FieldName}}
                {{$m.Method}}: mock.fn(),
                {{- end}}
             {{- end}}
        };
        {{- end}}

        {{- range $s.MocksSetup}}
        // Mock Return
        {{.Dependency}}.{{.Method}}.mock.mockImplementation(() => {{.ReturnValue | FormatValue}});
        {{- end}}

        // Init SUT
        // const sut = new {{$.Target.ClassName}}({{range $i, $e := $.Dependencies}}{{if $i}}, {{end}}{{$e.FieldName}}{{end}});

        // Act
        // const result = sut.{{$.Target.MethodName}}();

        // Assert
        {{- if $s.Expectations.ReturnValue}}
        // assert.strictEqual(result, {{$s.Expectations.ReturnValue | FormatValue}});
        {{- end}}
    });
    {{- end}}
});`

// TEMPLATE KOTLIN (MockK + JUnit5)
const kotlinTmpl = `import io.mockk.every
import io.mockk.mockk
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.Assertions.assertEquals

class {{.Target.ClassName}}Test {
    {{- range .Dependencies}}
    private val {{.FieldName}}: {{.InterfaceName}} = mockk()
    {{- end}}
    // private val sut = {{.Target.ClassName}}({{range $i, $e := .Dependencies}}{{if $i}}, {{end}}{{$e.FieldName}}{{end}})

    {{- if not .Scenarios}}
    @Test
    fun ` + "`should execute correctly`" + `() {
        // val result = sut.{{.Target.MethodName}}()
        // assertEquals(expected, result)
    }
    {{- end}}

    {{- range $s := .Scenarios}}
    @Test
    fun ` + "`{{$s.Description}}`" + `() {
        // Arrange
        {{- range $m := $s.MocksSetup}}
        every { {{.Dependency}}.{{.Method}}(any()) } returns {{.ReturnValue | FormatValue}}
        {{- end}}

        // Act
        // val result = sut.{{$.Target.MethodName}}()

        // Assert
        {{- if $s.Expectations.ReturnValue}}
        // assertEquals({{$s.Expectations.ReturnValue | FormatValue}}, result)
        {{- end}}
    }
    {{- end}}
}`

// TEMPLATE JAVA (Mockito + JUnit5)
const javaTmpl = `import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.InjectMocks;
import org.mockito.junit.jupiter.MockitoExtension;
import static org.mockito.Mockito.when;
import static org.mockito.ArgumentMatchers.any;
import static org.junit.jupiter.api.Assertions.assertEquals;

@ExtendWith(MockitoExtension.class)
class {{.Target.ClassName}}Test {

    {{- range .Dependencies}}
    @Mock
    {{.InterfaceName}} {{.FieldName}};
    {{- end}}

    @InjectMocks
    {{.Target.ClassName}} sut;

    {{- if not .Scenarios}}
    @Test
    void shouldExecuteCorrectly() {
        // var result = sut.{{.Target.MethodName}}();
        // assertEquals(expected, result);
    }
    {{- end}}

    {{- range $s := .Scenarios}}
    @Test
    void {{$s.ID | ToCamel}}() {
        // Arrange
        {{- range $m := $s.MocksSetup}}
        when({{.Dependency}}.{{.Method}}(any())).thenReturn({{.ReturnValue | FormatValue}});
        {{- end}}

        // Act
        // var result = sut.{{$.Target.MethodName}}();

        // Assert
        {{- if $s.Expectations.ReturnValue}}
        // assertEquals({{$s.Expectations.ReturnValue | FormatValue}}, result);
        {{- end}}
    }
    {{- end}}
}`

// TEMPLATE PHP (PHPUnit)
const phpTmpl = `<?php
use PHPUnit\Framework\TestCase;

class {{.Target.ClassName}}Test extends TestCase
{
    {{- if not .Scenarios}}
    public function testShouldExecuteCorrectly()
    {
        // $sut = new {{.Target.ClassName}}();
        // $this->assertTrue(true);
    }
    {{- end}}

    {{- range $s := .Scenarios}}
    public function test{{$s.ID | ToPascal}}()
    {
        // Arrange
        {{- range $.Dependencies}}
        ${{.FieldName}} = $this->createMock({{.InterfaceName}}::class);
        {{- end}}
        
        {{- range $m := $s.MocksSetup}}
        ${{.Dependency}}->method('{{.Method}}')->willReturn({{.ReturnValue | FormatValue}});
        {{- end}}

        // $sut = new {{$.Target.ClassName}}({{range $i, $e := $.Dependencies}}{{if $i}}, {{end}}${{$e.FieldName}}{{end}});

        // Act
        // $result = $sut->{{$.Target.MethodName}}();

        // Assert
        {{- if $s.Expectations.ReturnValue}}
        // $this->assertEquals({{$s.Expectations.ReturnValue | FormatValue}}, $result);
        {{- end}}
    }
    {{- end}}
}`

// TEMPLATE TYPESCRIPT (Jest)
const typeScriptTmpl = `//import { {{.Target.ClassName}} } from './{{.Target.ClassName}}';
import {describe, beforeEach, it, expect, test } from '@jest/globals';

describe('{{.Target.ClassName}}', () => {
    let sut: {{.Target.ClassName}};
    {{- range .Dependencies}}
    let {{.FieldName}}: any;
    {{- end}}

    beforeEach(() => {
        {{- range .Dependencies}}
        {{.FieldName}} = {
            // Mock methods here
        };
        {{- end}}
        // sut = new {{.Target.ClassName}}({{range $i, $e := .Dependencies}}{{if $i}}, {{end}}{{.FieldName}}{{end}});
    });

    {{- if not .Scenarios}}
    it('should work', () => {
        // const result = sut.{{.Target.MethodName}}();
        // expect(result).toBeDefined();
    });
    {{- end}}

    {{- range $s := .Scenarios}}
    it('{{$s.Description}}', () => {
        // Arrange
        {{- range $m := $s.MocksSetup}}
        {{.Dependency}}.{{.Method}} = jest.fn().mockReturnValue({{.ReturnValue | FormatValue}});
        {{- end}}

        // Act
        // const result = sut.{{$.Target.MethodName}}();

        // Assert
        {{- if $s.Expectations.ReturnValue}}
        // expect(result).toBe({{$s.Expectations.ReturnValue | FormatValue}});
        {{- end}}
    });
    {{- end}}
});`

// --- 3. FUNÇÕES DE AJUDA ---

// Processa um template para UMA linguagem específica
func processTemplate(config MetaFramework, lang string) (string, error) {
	var tmplString string
	switch strings.ToLower(lang) {
	case "go", "golang":
		tmplString = goTmpl
	case "csharp", "cs":
		tmplString = csharpTmpl
	case "kotlin", "kt":
		tmplString = kotlinTmpl
	case "typescript", "ts":
		tmplString = typeScriptTmpl
	case "node", "js", "javascript":
		tmplString = nodeNativeTmpl
	case "java":
		tmplString = javaTmpl
	case "php":
		tmplString = phpTmpl
	default:
		return "", fmt.Errorf("unsupported language: %s", lang)
	}

	funcMap := template.FuncMap{
		"ToPascal": func(s string) string {
			return strings.ReplaceAll(strings.Title(strings.ReplaceAll(s, "_", " ")), " ", "")
		},
		"ToLower": func(s string) string {
			return strings.ToLower(s)
		},
		"ToCamel": func(s string) string {
			t := strings.ReplaceAll(strings.Title(strings.ReplaceAll(s, "_", " ")), " ", "")
			if len(t) > 0 { return strings.ToLower(t[0:1]) + t[1:] }
			return ""
		},
		"FormatValue": func(v interface{}) string {
			switch val := v.(type) {
			case string:
				return fmt.Sprintf(`"%s"`, val)
			default:
				return fmt.Sprintf("%v", val)
			}
		},
	}

	t, err := template.New("code").Funcs(funcMap).Parse(tmplString)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, config); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Mapeamento de Extensões
var extensions = map[string]string{
	"go":         "go",
	"golang":     "go",
	"csharp":     "cs",
	"cs":         "cs",
	"kotlin":     "kt",
	"kt":         "kt",
	"java":       "java",
	"php":        "php",
	"typescript": "ts",
	"ts":         "ts",
	"node":       "test.js",
	"js":         "test.js",
	"javascript": "test.js",
}

// Helper para determinar nome do arquivo
func getFilename(className, lang string) string {
	ext, exists := extensions[strings.ToLower(lang)]
	if !exists { ext = "txt" }
	
	suffix := "Test"
	// Exceções de nomenclatura
	if strings.Contains(ext, "test.js") { suffix = "" }
	if ext == "go" { suffix = "_test" }
	
	return fmt.Sprintf("%s%s.%s", className, suffix, ext)
}

// Processa um arquivo de spec e gera N arquivos de teste
func processSpecFile(path string, outputDir string) ([]string, error) {
    data, readErr := os.ReadFile(path)
    if readErr != nil {
        return nil, fmt.Errorf("error reading file %s: %v", path, readErr)
    }
    
    var config MetaFramework
    if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
        return nil, fmt.Errorf("error parsing JSON in %s: %v", path, jsonErr)
    }

    // Determina lista de linguagens
    languages := config.Meta.Langs
    if len(languages) == 0 && config.Meta.Lang != "" {
        languages = []string{config.Meta.Lang}
    }
    
    if len(languages) == 0 {
        return nil, fmt.Errorf("no language specified in %s", path)
    }

    var generatedFiles []string

    for _, lang := range languages {
        // 1. Gera Código
        code, err := processTemplate(config, lang)
        if err != nil {
            fmt.Printf("⚠️ Warning: Skipping %s in %s: %v\n", lang, path, err)
            continue
        }

        // 2. Determina Nome
        filename := getFilename(config.Target.ClassName, lang)
        finalPath := filepath.Join(outputDir, filename)

        // 3. Salva
        if err := os.WriteFile(finalPath, []byte(code), 0644); err != nil {
            return generatedFiles, fmt.Errorf("error saving %s: %v", finalPath, err)
        }
        generatedFiles = append(generatedFiles, finalPath)
    }

    return generatedFiles, nil
}

// --- 4. MAIN ---

func main() {
    start := time.Now()

    // Flags
	fileFlag := flag.String("file", "", "Path to JSON spec file (supports wildcards like specs/*.json)")
	outFlag := flag.String("out", "", "Output filename (Only used in simple mode)")
	
    // Simple Mode Flags
	langFlag := flag.String("lang", "", "Language")
	classFlag := flag.String("class", "", "Class Name")
    
    printFlag := flag.Bool("print", false, "Print to console (Simple mode only)")

	flag.Parse()

    // Configura pasta de saída padrão
    const outputDir = "test"
    if !*printFlag {
        if err := os.MkdirAll(outputDir, 0755); err != nil {
            fmt.Printf("❌ Error creating directory: %v\n", err)
            os.Exit(1)
        }
    }

    // --- MODO 1: ARQUIVO(S) DE ESPECIFICAÇÃO ---
	if *fileFlag != "" {
        // Expande wildcards (ex: *.json)
        files, globErr := filepath.Glob(*fileFlag)
        if globErr != nil {
            fmt.Printf("❌ Error with file pattern: %v\n", globErr)
            os.Exit(1)
        }
        if len(files) == 0 {
            // Tenta usar o arquivo literal se o glob não achou nada (caso o usuário não use *)
            if _, err := os.Stat(*fileFlag); err == nil {
                files = []string{*fileFlag}
            } else {
                fmt.Printf("❌ No files found matching: %s\n", *fileFlag)
                os.Exit(1)
            }
        }

        fmt.Println("⚡ OrchAxon AutoTest v1.0 (Batch Mode)")
        totalGenerated := 0
        
        for _, file := range files {
            generated, err := processSpecFile(file, outputDir)
            if err != nil {
                fmt.Printf("❌ Failed to process %s: %v\n", file, err)
            } else {
                for _, f := range generated {
                    fmt.Printf("✓ Generated %s (from %s)\n", f, filepath.Base(file))
                }
                totalGenerated += len(generated)
            }
        }
        
        elapsed := time.Since(start)
        fmt.Printf("\n✨ Done! %d files generated in %.2fs\n", totalGenerated, elapsed.Seconds())
        return
    }

    // --- MODO 2: SIMPLE CLI FLAGS ---
	if *langFlag != "" && *classFlag != "" {
		fakeConfig := MetaFramework{
			Meta: MetaInfo{Lang: *langFlag},
			Target: TargetInfo{ClassName: *classFlag, MethodName: "MyMethod"},
			Scenarios: []Scenario{{ID: "should_work", Description: "Should return expected result"}},
		}
		
        // Gera Código
        code, err := processTemplate(fakeConfig, *langFlag)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            os.Exit(1)
        }

        if *printFlag {
            fmt.Println(code)
            return
        }

        // Salva
        finalOutput := *outFlag
        if finalOutput == "" {
            finalOutput = getFilename(*classFlag, *langFlag)
        }
        finalPath := filepath.Join(outputDir, finalOutput)

        if err := os.WriteFile(finalPath, []byte(code), 0644); err != nil {
            fmt.Printf("❌ Error saving: %v\n", err)
            os.Exit(1)
        }

        elapsed := time.Since(start)
        fmt.Println("⚡ OrchAxon AutoTest v1.0 (Simple Mode)")
        fmt.Printf("✓ Generated %s (%.2fs)\n", finalPath, elapsed.Seconds())
        return
	} 
    
    // --- HELP ---
    fmt.Println("❌ Usage:")
    fmt.Println("  Batch Mode:  autotest -file \"specs/*.json\"")
    fmt.Println("  Simple Mode: autotest -lang node -class User")
    os.Exit(1)
}