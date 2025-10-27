package config

import (
	"bytes"
	"log/slog"
	"testing"
)

// Test struct without any log tags
type SimpleConfig struct {
	Name    string
	Port    int
	Enabled bool
}

func (s SimpleConfig) LogValue() slog.Value {
	return structToLogValue(s)
}

// Test struct with log:"-" tags
type SecureConfig struct {
	PublicKey string
	SecretKey string `log:"-"`
	APIKey    string `log:"-"`
	Port      int
	DebugMode bool
}

func (s SecureConfig) LogValue() slog.Value {
	return structToLogValue(s)
}

// Test struct with various types
type ComplexConfig struct {
	Name        string
	Count       int
	Rate        float64
	Enabled     bool
	OptionalStr *string
	OptionalInt *int
	SecretToken string `log:"-"`
}

func (c ComplexConfig) LogValue() slog.Value {
	return structToLogValue(c)
}

func TestStructToLogValue_SimpleStruct(t *testing.T) {
	config := SimpleConfig{
		Name:    "test-service",
		Port:    8080,
		Enabled: true,
	}

	logValue := config.LogValue()
	attrs := logValue.Group()

	if len(attrs) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(attrs))
	}

	attrMap := make(map[string]slog.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	if val, ok := attrMap["name"]; !ok || val.String() != "test-service" {
		t.Errorf("name field incorrect: %v", val)
	}
	if val, ok := attrMap["port"]; !ok || val.Int64() != 8080 {
		t.Errorf("port field incorrect: %v", val)
	}
	if val, ok := attrMap["enabled"]; !ok || val.Bool() != true {
		t.Errorf("enabled field incorrect: %v", val)
	}
}

func TestStructToLogValue_WithExcludedFields(t *testing.T) {
	config := SecureConfig{
		PublicKey: "pk_12345",
		SecretKey: "sk_secret",
		APIKey:    "api_key_secret",
		Port:      3000,
		DebugMode: false,
	}

	logValue := config.LogValue()
	attrs := logValue.Group()

	// Should only have 3 attributes (PublicKey, Port, DebugMode)
	if len(attrs) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(attrs))
	}

	attrMap := make(map[string]slog.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	// Verify excluded fields are not present
	if _, ok := attrMap["secret_key"]; ok {
		t.Error("secret_key should be excluded from logging")
	}
	if _, ok := attrMap["api_key"]; ok {
		t.Error("api_key should be excluded from logging")
	}

	// Verify included fields are present
	if val, ok := attrMap["public_key"]; !ok || val.String() != "pk_12345" {
		t.Errorf("public_key should be present, got %v", val)
	}
	if val, ok := attrMap["port"]; !ok || val.Int64() != 3000 {
		t.Errorf("port should be present with value 3000, got %v", val)
	}
	if val, ok := attrMap["debug_mode"]; !ok || val.Bool() != false {
		t.Errorf("debug_mode should be present with value false, got %v", val)
	}
}

func TestStructToLogValue_ComplexTypes(t *testing.T) {
	optStr := "optional"
	optInt := 42

	config := ComplexConfig{
		Name:        "complex-service",
		Count:       100,
		Rate:        99.5,
		Enabled:     true,
		OptionalStr: &optStr,
		OptionalInt: &optInt,
		SecretToken: "should-not-appear",
	}

	logValue := config.LogValue()
	attrs := logValue.Group()

	// Should have 6 attributes (all except SecretToken)
	if len(attrs) != 6 {
		t.Errorf("expected 6 attributes, got %d", len(attrs))
	}

	attrMap := make(map[string]slog.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	// Verify secret is excluded
	if _, ok := attrMap["secret_token"]; ok {
		t.Error("secret_token should be excluded from logging")
	}

	// Verify all other fields are present with correct values
	if val, ok := attrMap["name"]; !ok || val.String() != "complex-service" {
		t.Errorf("name field incorrect: %v", val)
	}
	if val, ok := attrMap["count"]; !ok || val.Int64() != 100 {
		t.Errorf("count field incorrect: %v", val)
	}
	if val, ok := attrMap["rate"]; !ok || val.Float64() != 99.5 {
		t.Errorf("rate field incorrect: %v", val)
	}
	if val, ok := attrMap["enabled"]; !ok || val.Bool() != true {
		t.Errorf("enabled field incorrect: %v", val)
	}
	if val, ok := attrMap["optional_str"]; !ok || val.String() != "optional" {
		t.Errorf("optional_str field incorrect: %v", val)
	}
	if val, ok := attrMap["optional_int"]; !ok || val.Int64() != 42 {
		t.Errorf("optional_int field incorrect: %v", val)
	}
}

func TestStructToLogValue_NilPointers(t *testing.T) {
	config := ComplexConfig{
		Name:        "nil-test",
		Count:       1,
		Rate:        1.0,
		Enabled:     false,
		OptionalStr: nil,
		OptionalInt: nil,
		SecretToken: "secret",
	}

	logValue := config.LogValue()
	attrs := logValue.Group()

	// Should have 4 attributes (nil pointers and SecretToken excluded)
	if len(attrs) != 4 {
		t.Errorf("expected 4 attributes, got %d", len(attrs))
	}

	attrMap := make(map[string]slog.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	// Verify nil pointers are not present
	if _, ok := attrMap["optional_str"]; ok {
		t.Error("optional_str (nil) should not be present")
	}
	if _, ok := attrMap["optional_int"]; ok {
		t.Error("optional_int (nil) should not be present")
	}
}

func TestStructToLogValue_ActualLogOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	config := SecureConfig{
		PublicKey: "pk_test",
		SecretKey: "sk_should_not_appear",
		APIKey:    "api_should_not_appear",
		Port:      9000,
		DebugMode: true,
	}

	logger.Info("config loaded", "config", config)

	output := buf.String()

	// Verify public key is in output
	if !bytes.Contains(buf.Bytes(), []byte("pk_test")) {
		t.Error("public_key should appear in log output")
	}

	// Verify secrets are NOT in output
	if bytes.Contains(buf.Bytes(), []byte("sk_should_not_appear")) {
		t.Error("secret_key should NOT appear in log output")
	}
	if bytes.Contains(buf.Bytes(), []byte("api_should_not_appear")) {
		t.Error("api_key should NOT appear in log output")
	}

	// Verify other fields are present
	if !bytes.Contains(buf.Bytes(), []byte("9000")) {
		t.Error("port should appear in log output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("true")) {
		t.Error("debug_mode should appear in log output")
	}

	t.Logf("Log output: %s", output)
}

func TestStructToLogValue_EmptyStruct(t *testing.T) {
	type EmptyConfig struct{}

	empty := EmptyConfig{}
	logValue := structToLogValue(empty)
	attrs := logValue.Group()

	if len(attrs) != 0 {
		t.Errorf("expected 0 attributes for empty struct, got %d", len(attrs))
	}
}

func TestStructToLogValue_AllFieldsExcluded(t *testing.T) {
	type AllSecretConfig struct {
		Secret1 string `log:"-"`
		Secret2 int    `log:"-"`
		Secret3 bool   `log:"-"`
	}

	config := AllSecretConfig{
		Secret1: "secret",
		Secret2: 123,
		Secret3: true,
	}

	logValue := structToLogValue(config)
	attrs := logValue.Group()

	if len(attrs) != 0 {
		t.Errorf("expected 0 attributes when all fields excluded, got %d", len(attrs))
	}
}

func TestStructToLogValue_MixedIntTypes(t *testing.T) {
	type IntTypesConfig struct {
		Int8Val   int8
		Int16Val  int16
		Int32Val  int32
		Int64Val  int64
		UintVal   uint
		Uint8Val  uint8
		Uint16Val uint16
		Uint32Val uint32
		Uint64Val uint64
	}

	config := IntTypesConfig{
		Int8Val:   -8,
		Int16Val:  -16,
		Int32Val:  -32,
		Int64Val:  -64,
		UintVal:   1,
		Uint8Val:  8,
		Uint16Val: 16,
		Uint32Val: 32,
		Uint64Val: 64,
	}

	logValue := structToLogValue(config)
	attrs := logValue.Group()

	if len(attrs) != 9 {
		t.Errorf("expected 9 attributes, got %d", len(attrs))
	}

	attrMap := make(map[string]slog.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	// Verify signed ints
	if val, ok := attrMap["int8_val"]; !ok || val.Int64() != -8 {
		t.Errorf("int8_val incorrect: %v", val)
	}
	if val, ok := attrMap["int64_val"]; !ok || val.Int64() != -64 {
		t.Errorf("int64_val incorrect: %v", val)
	}

	// Verify unsigned ints
	if val, ok := attrMap["uint8_val"]; !ok || val.Uint64() != 8 {
		t.Errorf("uint8_val incorrect: %v", val)
	}
	if val, ok := attrMap["uint64_val"]; !ok || val.Uint64() != 64 {
		t.Errorf("uint64_val incorrect: %v", val)
	}
}

func TestStructToLogValue_FloatTypes(t *testing.T) {
	type FloatConfig struct {
		Float32Val float32
		Float64Val float64
	}

	config := FloatConfig{
		Float32Val: 3.14,
		Float64Val: 2.71828,
	}

	logValue := structToLogValue(config)
	attrs := logValue.Group()

	if len(attrs) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(attrs))
	}

	attrMap := make(map[string]slog.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	if val, ok := attrMap["float32_val"]; !ok {
		t.Error("float32_val should be present")
	} else {
		// float32 may lose precision when converted to float64
		if val.Float64() < 3.13 || val.Float64() > 3.15 {
			t.Errorf("float32_val incorrect: %v", val)
		}
	}

	if val, ok := attrMap["float64_val"]; !ok || val.Float64() != 2.71828 {
		t.Errorf("float64_val incorrect: %v", val)
	}
}
