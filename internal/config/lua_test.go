package config

import (
	"reflect"
	"strings"
	"testing"
)

type SimplePrimitives struct {
	Name   string
	Age    int
	Active bool
	Score  float64
}

type WithSlice struct {
	Tags []string
}

type WithStringMap struct {
	Metadata map[string]string
}

type WithIntMap struct {
	Counts map[int]string
}

type CamelCaseFields struct {
	FirstName string
	LastName  string
	UserID    int
}

type AllNumericTypes struct {
	Int8Field    int8
	Int16Field   int16
	Int32Field   int32
	Int64Field   int64
	Uint8Field   uint8
	Uint16Field  uint16
	Uint32Field  uint32
	Uint64Field  uint64
	Float32Field float32
	Float64Field float64
}

type NestedSlice struct {
	Matrix [][]int
}

type ComplexMap struct {
	Data map[string][]int
}

type User struct {
	Name  string
	Email string
	Age   int
}

type WithUnexported struct {
	PublicField   string
	AnotherPublic int
}

type WithPointer struct {
	Name *string
	Age  *int
}

func TestGenerateLuaType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
		wantErr  bool
	}{
		{
			name:  "simple struct with primitives",
			input: SimplePrimitives{},
			expected: []string{
				"---@class SimplePrimitives",
				"---@field name string",
				"---@field age number",
				"---@field active boolean",
				"---@field score number",
			},
			wantErr: false,
		},
		{
			name:  "struct with slice",
			input: WithSlice{},
			expected: []string{
				"---@class WithSlice",
				"---@field tags string[]",
			},
			wantErr: false,
		},
		{
			name:  "struct with map string keys",
			input: WithStringMap{},
			expected: []string{
				"---@class WithStringMap",
				"---@field metadata { [string]:string }",
			},
			wantErr: false,
		},
		{
			name:  "struct with map non-string keys",
			input: WithIntMap{},
			expected: []string{
				"---@class WithIntMap",
				"---@field counts table<number, string>",
			},
			wantErr: false,
		},
		{
			name:  "struct with CamelCase field names",
			input: CamelCaseFields{},
			expected: []string{
				"---@class CamelCaseFields",
				"---@field first_name string",
				"---@field last_name string",
				"---@field user_id number",
			},
			wantErr: false,
		},
		{
			name:  "struct with all numeric types",
			input: AllNumericTypes{},
			expected: []string{
				"---@class AllNumericTypes",
				"---@field int8_field number",
				"---@field int16_field number",
				"---@field int32_field number",
				"---@field int64_field number",
				"---@field uint8_field number",
				"---@field uint16_field number",
				"---@field uint32_field number",
				"---@field uint64_field number",
				"---@field float32_field number",
				"---@field float64_field number",
			},
			wantErr: false,
		},
		{
			name:  "struct with nested slice",
			input: NestedSlice{},
			expected: []string{
				"---@class NestedSlice",
				"---@field matrix number[][]",
			},
			wantErr: false,
		},
		{
			name:  "struct with complex map",
			input: ComplexMap{},
			expected: []string{
				"---@class ComplexMap",
				"---@field data { [string]:number[] }",
			},
			wantErr: false,
		},
		{
			name:  "named User struct",
			input: User{},
			expected: []string{
				"---@class User",
				"---@field name string",
				"---@field email string",
				"---@field age number",
			},
			wantErr: false,
		},
		{
			name:  "struct with pointers",
			input: WithPointer{},
			expected: []string{
				"---@class WithPointer",
				"---@field name? string",
				"---@field age? number",
			},
			wantErr: false,
		},
		{
			name:  "struct with unexported fields",
			input: WithUnexported{},
			expected: []string{
				"---@class WithUnexported",
				"---@field public_field string",
				"---@field another_public number",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateLuaType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateLuaType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotLines := strings.Split(got, "\n")
				if len(gotLines) != len(tt.expected) {
					t.Errorf("GenerateLuaType() got %d lines, expected %d\nGot:\n%s\nExpected:\n%s",
						len(gotLines), len(tt.expected), got, strings.Join(tt.expected, "\n"))
					return
				}
				for i, line := range gotLines {
					if line != tt.expected[i] {
						t.Errorf("GenerateLuaType() line %d:\ngot:      %q\nexpected: %q", i, line, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"lastName", "last_name"},
		{"UserID", "user_id"},
		{"HTTPSConnection", "https_connection"},
		{"simple", "simple"},
		{"", ""},
		{"A", "a"},
		{"AB", "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToLowerSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"LastName", "last_name"},
		{"UserID", "user_id"},
		{"HTTPSConnection", "https_connection"},
		{"simple", "simple"},
		{"", ""},
		{"A", "a"},
		{"AB", "ab"},
		{"DatabasePath", "database_path"},
		{"OauthClientCredentials", "oauth_client_credentials"},
		{"redirectURL", "redirect_url"},
		{"parseHTMLFile", "parse_html_file"},
		{"XMLParser", "xml_parser"},
		{"APIKey", "api_key"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToLowerSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("ToLowerSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		snakeCase string
		camelCase string
	}{
		{"name", "Name"},
		{"first_name", "FirstName"},
		{"database_path", "DatabasePath"},
	}

	for _, tt := range tests {
		t.Run(tt.snakeCase, func(t *testing.T) {
			// snake_case -> UpperCamelCase
			gotCamel := ToUpperCamelCase(tt.snakeCase)
			if gotCamel != tt.camelCase {
				t.Errorf("ToUpperCamelCase(%q) = %q, want %q", tt.snakeCase, gotCamel, tt.camelCase)
			}

			// UpperCamelCase -> snake_case
			gotSnake := ToLowerSnakeCase(tt.camelCase)
			if gotSnake != tt.snakeCase {
				t.Errorf("ToLowerSnakeCase(%q) = %q, want %q", tt.camelCase, gotSnake, tt.snakeCase)
			}
		})
	}
}

func TestLuaType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
	}{
		{"string type", "", "string", false},
		{"int type", 0, "number", false},
		{"bool type", false, "boolean", false},
		{"slice of strings", []string{}, "string[]", false},
		{"slice of ints", []int{}, "number[]", false},
		{"map with string keys", map[string]int{}, "{ [string]:number }", false},
		{"map with int keys", map[int]string{}, "table<number, string>", false},
		{"nested slice", [][]string{}, "string[][]", false},
		{"pointer to string", new(string), "string?", false},
		{"pointer to int", new(int), "number?", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeData, err := luaType(reflect.TypeOf(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("luaType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				typ := typeData.String()
				if typ != tt.expected {
					t.Errorf("luaType() = %q, want %q", typ, tt.expected)
				}
			}
		})
	}
}
