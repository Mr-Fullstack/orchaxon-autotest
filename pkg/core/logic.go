package core

import (
	"fmt"
	"strings"
	"text/template"
)

// --- ESTRUTURAS DE DADOS (Igual ao seu original) ---

type MetaFramework struct {
	Meta         MetaInfo     `json:"meta"`
	Target       TargetInfo   `json:"target"`
	Dependencies []Dependency `json:"dependencies"`
	Scenarios    []Scenario   `json:"scenarios"`
}

type MetaInfo struct {
	Lang  string   `json:"lang"`
	Langs []string `json:"langs"`
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


const pythonTmpl = `import unittest
from unittest.mock import MagicMock

class Test{{.Target.ClassName}}(unittest.TestCase):
    def setUp(self):
        {{- range .Dependencies}}
        self.mock_{{.FieldName}} = MagicMock()
        {{- end}}
        # Assumes constructor injection
        # self.sut = {{.Target.ClassName}}({{range $i, $e := .Dependencies}}{{if $i}}, {{end}}self.mock_{{.FieldName}}{{end}})

    {{- range $s := .Scenarios}}
    def test_{{$s.ID | ToSnake}}(self):
        """ {{$s.Description}} """
        # Arrange
        {{- range $m := $s.MocksSetup}}
        self.mock_{{.Dependency}}.{{.Method}}.ReturnValue = {{.ReturnValue | FormatValue}}
        {{- end}}

        # Act
        # result = self.sut.{{$.Target.MethodName}}()

        # Assert
        {{- if $s.Expectations.ReturnValue}}
        # self.assertEqual(result, {{$s.Expectations.ReturnValue | FormatValue}})
        {{- end}}
    {{- end}}` 


const nodeTmpl = `import { describe, it, mock } from 'node:test';
import assert from 'node:assert';
// import { {{.Target.ClassName}} } from '../src/{{.Target.ClassName}}.js'; 

describe('{{.Target.ClassName}}', () => {
    
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
        // Configure Mock Return
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

// --- FUNÇÃO CORE (Lógica Pura) ---

func ProcessTemplate(config MetaFramework, lang string) (string, error) {
	var tmplString string
	switch strings.ToLower(lang) {
	case "go", "golang":
		tmplString = goTmpl
	case "csharp", "cs","c#":
		tmplString = csharpTmpl
	case "kotlin", "kt":
		tmplString = kotlinTmpl
	case "java":
		tmplString = javaTmpl
	case "typescript", "ts":
		tmplString = typeScriptTmpl
	case "python", "py":
		tmplString = pythonTmpl
	case "php":
		tmplString = phpTmpl
    case "node", "js", "javascript":
		tmplString = nodeTmpl
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
			if len(t) > 0 {
				return strings.ToLower(t[0:1]) + t[1:]
			}
			return ""
		},
        "ToSnake": func(s string) string {
            return strings.ReplaceAll(strings.ToLower(s), " ", "_")
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