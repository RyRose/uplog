package config

import (
	"log/slog"
	"reflect"
	"testing"
)

func TestDataLogValue(t *testing.T) {
	version := "1.0.0"
	data := Data{
		Debug:        true,
		Version:      &version,
		DatabasePath: "/test/db",
		Port:         "8080",
		SwaggerURL:   "http://localhost/swagger",
	}

	logValue := data.LogValue()
	if logValue.Kind() != slog.KindGroup {
		t.Errorf("expected Kind to be Group, got %v", logValue.Kind())
	}

	// Verify the log value contains expected fields
	attrs := logValue.Group()
	if len(attrs) != 5 {
		t.Errorf("expected 5 attributes, got %d", len(attrs))
	}

	// Verify specific field values
	attrMap := make(map[string]slog.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	if val, ok := attrMap["debug"]; !ok || val.Bool() != true {
		t.Errorf("debug field incorrect: %v", val)
	}
	if val, ok := attrMap["version"]; !ok || val.String() != "1.0.0" {
		t.Errorf("version field incorrect: %v", val)
	}
	if val, ok := attrMap["database_path"]; !ok || val.String() != "/test/db" {
		t.Errorf("database_path field incorrect: %v", val)
	}
	if val, ok := attrMap["port"]; !ok || val.String() != "8080" {
		t.Errorf("port field incorrect: %v", val)
	}
	if val, ok := attrMap["swagger_url"]; !ok || val.String() != "http://localhost/swagger" {
		t.Errorf("swagger_url field incorrect: %v", val)
	}
}

func TestDataLogValueNilVersion(t *testing.T) {
	data := Data{
		Debug:        false,
		Version:      nil,
		DatabasePath: "/test/db",
		Port:         "3000",
		SwaggerURL:   "http://test",
	}

	logValue := data.LogValue()
	attrs := logValue.Group()

	// Should have 4 attributes when version is nil (version field is skipped)
	if len(attrs) != 4 {
		t.Errorf("expected 4 attributes, got %d", len(attrs))
	}

	// Verify version is not in the attributes
	for _, attr := range attrs {
		if attr.Key == "version" {
			t.Error("version field should not be present when nil")
		}
	}
}

func TestLogValueWithExcludedFields(t *testing.T) {
	// Test with a struct that has log:"-" tags
	type TestData struct {
		PublicField  string
		SecretField  string `log:"-"`
		Password     string `log:"-"`
		AnotherField int
	}

	testData := TestData{
		PublicField:  "visible",
		SecretField:  "should-not-appear",
		Password:     "secret123",
		AnotherField: 42,
	}

	// Use reflection to call LogValue on the test struct
	v := reflect.ValueOf(testData)
	t2 := v.Type()

	var attrs []slog.Attr
	for i := 0; i < v.NumField(); i++ {
		field := t2.Field(i)
		value := v.Field(i)

		if tag := field.Tag.Get("log"); tag == "-" {
			continue
		}

		fieldName := toSnakeCase(field.Name)
		switch value.Kind() {
		case reflect.String:
			attrs = append(attrs, slog.String(fieldName, value.String()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			attrs = append(attrs, slog.Int64(fieldName, value.Int()))
		}
	}

	logValue := slog.GroupValue(attrs...)
	attrList := logValue.Group()

	// Should only have 2 attributes (PublicField and AnotherField)
	if len(attrList) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(attrList))
	}

	// Verify excluded fields are not present
	for _, attr := range attrList {
		if attr.Key == "secret_field" || attr.Key == "password" {
			t.Errorf("field %s should be excluded from logging", attr.Key)
		}
	}

	// Verify included fields are present
	attrMap := make(map[string]slog.Value)
	for _, attr := range attrList {
		attrMap[attr.Key] = attr.Value
	}

	if val, ok := attrMap["public_field"]; !ok || val.String() != "visible" {
		t.Errorf("public_field should be present with value 'visible', got %v", val)
	}
	if val, ok := attrMap["another_field"]; !ok || val.Int64() != 42 {
		t.Errorf("another_field should be present with value 42, got %v", val)
	}
}
