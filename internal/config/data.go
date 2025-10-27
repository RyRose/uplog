package config

import (
	"log/slog"
	"reflect"
)

type Data struct {
	Debug        bool
	Version      *string
	DatabasePath string
	Port         string
	SwaggerURL   string
}

func (u Data) LogValue() slog.Value {
	return structToLogValue(u)
}

// structToLogValue converts a struct to slog.Value using reflection.
// Fields with `log:"-"` tag are excluded from the output.
// Field names are automatically converted to snake_case.
func structToLogValue(v interface{}) slog.Value {
	var attrs []slog.Attr

	val := reflect.ValueOf(v)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// Check if field has log:"-" tag to skip logging
		if tag := field.Tag.Get("log"); tag == "-" {
			continue
		}

		// Get the field name for logging (convert to snake_case)
		fieldName := toSnakeCase(field.Name)

		// Convert value to slog.Attr
		if attr, ok := valueToAttr(fieldName, value); ok {
			attrs = append(attrs, attr)
		}
	}

	return slog.GroupValue(attrs...)
}

// valueToAttr converts a reflect.Value to slog.Attr.
// Returns (attr, false) if the value should be skipped (e.g., nil pointer).
func valueToAttr(name string, value reflect.Value) (slog.Attr, bool) {
	switch value.Kind() {
	case reflect.Bool:
		return slog.Bool(name, value.Bool()), true
	case reflect.String:
		return slog.String(name, value.String()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return slog.Int64(name, value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return slog.Uint64(name, value.Uint()), true
	case reflect.Float32, reflect.Float64:
		return slog.Float64(name, value.Float()), true
	case reflect.Ptr:
		if value.IsNil() {
			return slog.Attr{}, false
		}
		// Dereference pointer and handle the underlying value
		return valueToAttr(name, value.Elem())
	default:
		// For any other types, use slog.Any
		return slog.Any(name, value.Interface()), true
	}
}

// LuaTypes returns a slice of config types for Lua integration.
func LuaTypes() []any {
	return []any{
		Data{},
	}
}
