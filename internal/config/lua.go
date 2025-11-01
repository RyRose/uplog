package config

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// ToLowerSnakeCase converts from UpperCamelCase to lower_snake_case.
// This is the inverse of ToUpperCamelCase.
// Handles acronyms properly: redirectURL -> redirect_url, UserID -> user_id
func ToLowerSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	var result []rune
	runes := []rune(s)

	for i := range len(runes) {
		r := runes[i]

		if unicode.IsUpper(r) {
			// Add underscore before uppercase letter if:
			// 1. Not at the beginning
			// 2. Previous character is lowercase OR
			// 3. Next character exists and is lowercase (end of acronym)
			if i > 0 {
				prevLower := unicode.IsLower(runes[i-1])
				nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

				if prevLower || nextLower {
					result = append(result, '_')
				}
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func toSnakeCase(s string) string {
	return ToLowerSnakeCase(s)
}

// ToUpperCamelCase converts strings from snake_case to UpperCamelCase.
func ToUpperCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	var result []rune
	capitalizeNext := true

	for _, r := range s {
		if r == '_' {
			capitalizeNext = true
		} else {
			if capitalizeNext {
				result = append(result, unicode.ToUpper(r))
				capitalizeNext = false
			} else {
				result = append(result, r)
			}
		}
	}

	return string(result)
}

type luaTypeData struct {
	value    string
	optional bool
}

func (l *luaTypeData) String() string {
	if l.optional {
		return l.value + "?"
	}
	return l.value
}

func luaType(t reflect.Type) (*luaTypeData, error) {
	switch t.Kind() {
	case reflect.String:
		return &luaTypeData{value: "string"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return &luaTypeData{value: "number"}, nil
	case reflect.Bool:
		return &luaTypeData{value: "boolean"}, nil
	case reflect.Slice:
		s, err := luaType(t.Elem())
		if err != nil {
			return nil, err
		}
		return &luaTypeData{value: s.String() + "[]"}, nil
	// TODO: Verify maps are handled as expected in gluamapper.
	case reflect.Map:
		keyS, err := luaType(t.Key())
		if err != nil {
			return nil, fmt.Errorf("map key type: %w", err)
		}
		elemS, err := luaType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("map value type: %w", err)
		}
		if keyS.value == "string" {
			return &luaTypeData{value: "{ [string]:" + elemS.String() + " }"}, nil
		}
		return &luaTypeData{value: "table<" + keyS.String() + ", " + elemS.String() + ">"}, nil
	case reflect.Struct:
		return &luaTypeData{value: t.Name()}, nil
	case reflect.Pointer:
		s, err := luaType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("pointer to %v: %w", t.Elem(), err)
		}
		return &luaTypeData{value: s.String(), optional: true}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %v", t)
	}
}

func GenerateLuaType(v any) (string, error) {
	t := reflect.TypeOf(v)
	lines := []string{fmt.Sprintf("---@class %s", t.Name())}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Skip unexported (lowercase) fields as they are hidden from Lua
		if len(f.Name) > 0 && !unicode.IsUpper(rune(f.Name[0])) {
			continue
		}

		luaTypeData, err := luaType(f.Type)
		if err != nil {
			return "", fmt.Errorf("lines so far:\n%s\nfield %s: %w", strings.Join(lines, "\n"), f.Name, err)
		}
		fieldName := toSnakeCase(f.Name)
		if luaTypeData.optional {
			fieldName += "?"
		}
		lines = append(lines, fmt.Sprintf("---@field %s %s", fieldName, luaTypeData.value))
	}
	return strings.Join(lines, "\n"), nil
}

func getZeroValues(v any) []any {
	seen := make(map[reflect.Type]bool)
	return getZeroValuesWithSeen(v, seen)
}

func getZeroValuesWithSeen(v any, seen map[reflect.Type]bool) []any {
	var zeroValues []any
	val := reflect.ValueOf(v)

	// Ensure we are working with the value of the struct
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return zeroValues
	}

	// Skip if we've already seen this type
	t := val.Type()
	if seen[t] {
		return zeroValues
	}
	seen[t] = true

	// Add the zero value of the current struct
	zeroValues = append(zeroValues, reflect.New(t).Elem().Interface())

	// Iterate through fields to find nested structs
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// Check if the field is a struct or a pointer to a struct
		if field.Kind() == reflect.Struct {
			zeroValues = append(zeroValues, getZeroValuesWithSeen(field.Interface(), seen)...)
		} else if field.Kind() == reflect.Pointer && fieldType.Type.Elem().Kind() == reflect.Struct {
			zeroValues = append(zeroValues, getZeroValuesWithSeen(reflect.New(fieldType.Type.Elem()).Elem().Interface(), seen)...)
		}
	}

	return zeroValues
}

// LuaTypes returns a slice of config types for Lua integration.
func LuaTypes() []any {
	return getZeroValues(Data{})
}
