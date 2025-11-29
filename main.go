package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath" // <--- NOVO: Para lidar com pastas em Windows/Linux
	"strings"
	"text/template"
)

// --- 1. ESTRUTURAS JSON (Mantidas iguais) ---
type MetaFramework struct {
	Meta         MetaInfo     `json:"meta"`
	Target       TargetInfo   `json:"target"`
	Dependencies []Dependency `json:"dependencies"`
	Scenarios    []Scenario   `json:"scenarios"`
}
type MetaInfo struct {
	Lang string `json:"lang"`
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
	ID           string      `json:"id"`
	Description  string      `json:"description"`
	MocksSetup   []MockSetup `json:"mocks_setup"`
	Expectations Expectation `json:"expectations"`
}
type MockSetup struct {
	Dependency  string      `json:"dependency"`
	Method      string      `json:"method"`
	ReturnValue interface{} `json:"return_value"`
}
type Expectation struct {
	ReturnValue interface{} `json:"return_value"`
}

// --- 2. TEMPLATES (Mantidos iguais) ---

const nodeNativeTmpl = `import { describe, it, mock } from 'node:test';
import assert from 'node:assert';
// import { {{.Target.ClassName}} } from '../src/{{.Target.ClassName}}.js'; 

describe('{{.Target.ClassName}}', () => {
    
    {{- if not .Scenarios}}
    it('should execute correctly', () => {
        // const sut = new {{.Target.ClassName}}();
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

// (Outros templates omitidos para brevidade, mas o código suporta todos se você colar aqui)

// --- 3. EXECUÇÃO ---

// Mapeamento de Extensões
var extensions = map[string]string{
	"csharp":     "cs",
	"cs":         "cs",
	"kotlin":     "kt",
	"kt":         "kt",
	"java":       "java",
	"php":        "php",
	"typescript": "ts",
	"ts":         "ts",
	"node":       "test.js", // Padrão Node Nativo
	"js":         "test.js",
	"javascript": "test.js",
}

func main() {
	fileFlag := flag.String("file", "", "Path to JSON spec file")
	outFlag := flag.String("out", "", "Output filename (Optional - Auto-generated if empty)")
	langFlag := flag.String("lang", "", "Language")
	classFlag := flag.String("class", "", "Class Name")
    // Adicionei uma flag para forçar a saída na tela se quiser
    printFlag := flag.Bool("print", false, "Print to console instead of saving")

	flag.Parse()

	var code string
	var err error
    var lang string
    var className string

	// 1. GERAÇÃO DO CÓDIGO
	if *fileFlag != "" {
        // Modo JSON: Precisamos ler o JSON para saber a classe/lang
		data, _ := os.ReadFile(*fileFlag)
		var config MetaFramework
		json.Unmarshal(data, &config)
        
        lang = config.Meta.Lang
        className = config.Target.ClassName
		code, err = processTemplate(config)

	} else if *langFlag != "" && *classFlag != "" {
        // Modo Simples
        lang = *langFlag
        className = *classFlag
		fakeConfig := MetaFramework{
			Meta: MetaInfo{Lang: *langFlag},
			Target: TargetInfo{ClassName: *classFlag, MethodName: "MyMethod"},
			Scenarios: []Scenario{{ID: "should_work", Description: "Should return expected result"}},
		}
		code, err = processTemplate(fakeConfig)
	} else {
		fmt.Println("❌ Usage: autotest -lang node -class User")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

    // 2. DECISÃO: SALVAR OU IMPRIMIR?
    if *printFlag {
        fmt.Println(code)
        return
    }

    // 3. AUTO-NAMING (A Mágica Acontece Aqui)
    finalOutput := *outFlag
    if finalOutput == "" {
        // Se o usuário não passou -out, a gente inventa
        ext, exists := extensions[strings.ToLower(lang)]
        if !exists {
            ext = "txt" // Fallback
        }
        
        // Padrão: UserTest.cs ou User.test.js
        suffix := "Test"
        if strings.Contains(ext, "test.js") {
            suffix = "" // Node já tem .test no nome
        }
        
        finalOutput = fmt.Sprintf("%s%s.%s", className, suffix, ext)
    }

	// 4. SALVAR NA PASTA TEST
    const outputDir = "test"
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        fmt.Printf("❌ Error creating directory: %v\n", err)
        os.Exit(1)
    }

    finalPath := filepath.Join(outputDir, finalOutput)

    saveErr := os.WriteFile(finalPath, []byte(code), 0644)
    if saveErr != nil {
        fmt.Printf("❌ Error saving file: %v\n", saveErr)
        os.Exit(1)
    }
    
    fmt.Printf("✅ Generated: %s\n", finalPath)
}

// ... (O resto das funções auxiliares continua igual) ...

// --- HELPERS ---

func generateFromJSON(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil { return "", err }
	var config MetaFramework
	if err := json.Unmarshal(data, &config); err != nil { return "", err }
	return processTemplate(config)
}

func processTemplate(config MetaFramework) (string, error) {
	var tmplString string
	switch strings.ToLower(config.Meta.Lang) {
	case "node", "javascript", "js": tmplString = nodeNativeTmpl
	case "csharp", "cs": tmplString = csharpTmpl
	default: return "", fmt.Errorf("unsupported language: %s", config.Meta.Lang)
	}

	funcMap := template.FuncMap{
		"ToPascal": func(s string) string { return strings.ReplaceAll(strings.Title(strings.ReplaceAll(s, "_", " ")), " ", "") },
		"ToCamel": func(s string) string { t := strings.ReplaceAll(strings.Title(strings.ReplaceAll(s, "_", " ")), " ", ""); if len(t) > 0 { return strings.ToLower(t[0:1]) + t[1:] }; return "" },
		"FormatValue": func(v interface{}) string {
			switch val := v.(type) {
			case string: return fmt.Sprintf(`"%s"`, val)
			default: return fmt.Sprintf("%v", val)
			}
		},
	}

	t, err := template.New("code").Funcs(funcMap).Parse(tmplString)
	if err != nil { return "", err }
	var buf strings.Builder
	if err := t.Execute(&buf, config); err != nil { return "", err }
	return buf.String(), nil
}