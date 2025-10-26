package util

import (
	"testing"
)

func TestUrlPathJoin(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		parts    []string
		expected string
	}{
		{
			name:     "simple join",
			base:     "/api",
			parts:    []string{"users", "123"},
			expected: "/api/users/123",
		},
		{
			name:     "with special characters",
			base:     "/api",
			parts:    []string{"users", "john?name=test"},
			expected: "/api/users/john%3Fname=test",
		},
		{
			name:     "with spaces",
			base:     "/api",
			parts:    []string{"users", "john doe"},
			expected: "/api/users/john%20doe",
		},
		{
			name:     "empty parts",
			base:     "/api",
			parts:    []string{},
			expected: "/api",
		},
		{
			name:     "single part",
			base:     "/api",
			parts:    []string{"users"},
			expected: "/api/users",
		},
		{
			name:     "slash in part",
			base:     "/api",
			parts:    []string{"users/123"},
			expected: "/api/users%2F123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UrlPathJoin(tt.base, tt.parts...)
			if got != tt.expected {
				t.Errorf("UrlPathJoin() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMinimum(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a is smaller", 1, 2, 1},
		{"b is smaller", 5, 3, 3},
		{"equal", 4, 4, 4},
		{"negative numbers", -5, -2, -5},
		{"zero and positive", 0, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Minimum(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("Minimum(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestMinimumFloat(t *testing.T) {
	got := Minimum(1.5, 2.7)
	if got != 1.5 {
		t.Errorf("Minimum(1.5, 2.7) = %f, want 1.5", got)
	}
}

func TestMaximum(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a is larger", 5, 2, 5},
		{"b is larger", 3, 7, 7},
		{"equal", 4, 4, 4},
		{"negative numbers", -5, -2, -2},
		{"zero and negative", 0, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Maximum(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("Maximum(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestMaximumString(t *testing.T) {
	got := Maximum("apple", "banana")
	if got != "banana" {
		t.Errorf("Maximum(\"apple\", \"banana\") = %q, want \"banana\"", got)
	}
}

func TestZero(t *testing.T) {
	t.Run("nil pointer returns zero value", func(t *testing.T) {
		var nilInt *int
		got := Zero(nilInt)
		if got != 0 {
			t.Errorf("Zero(nil *int) = %d, want 0", got)
		}
	})

	t.Run("pointer to value returns that value", func(t *testing.T) {
		val := 42
		got := Zero(&val)
		if got != 42 {
			t.Errorf("Zero(&42) = %d, want 42", got)
		}
	})

	t.Run("nil string pointer returns empty string", func(t *testing.T) {
		var nilStr *string
		got := Zero(nilStr)
		if got != "" {
			t.Errorf("Zero(nil *string) = %q, want \"\"", got)
		}
	})

	t.Run("pointer to string returns that string", func(t *testing.T) {
		val := "hello"
		got := Zero(&val)
		if got != "hello" {
			t.Errorf("Zero(&\"hello\") = %q, want \"hello\"", got)
		}
	})
}

func TestDeZero(t *testing.T) {
	t.Run("zero value returns nil", func(t *testing.T) {
		got := DeZero(0)
		if got != nil {
			t.Errorf("DeZero(0) = %v, want nil", got)
		}
	})

	t.Run("non-zero value returns pointer", func(t *testing.T) {
		got := DeZero(42)
		if got == nil {
			t.Fatal("DeZero(42) returned nil")
		}
		if *got != 42 {
			t.Errorf("*DeZero(42) = %d, want 42", *got)
		}
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		got := DeZero("")
		if got != nil {
			t.Errorf("DeZero(\"\") = %v, want nil", got)
		}
	})

	t.Run("non-empty string returns pointer", func(t *testing.T) {
		got := DeZero("hello")
		if got == nil {
			t.Fatal("DeZero(\"hello\") returned nil")
		}
		if *got != "hello" {
			t.Errorf("*DeZero(\"hello\") = %q, want \"hello\"", *got)
		}
	})
}

func TestZeroDeZeroRoundTrip(t *testing.T) {
	t.Run("int round trip", func(t *testing.T) {
		original := 42
		ptr := DeZero(original)
		got := Zero(ptr)
		if got != original {
			t.Errorf("round trip failed: got %d, want %d", got, original)
		}
	})

	t.Run("zero int returns nil then zero", func(t *testing.T) {
		original := 0
		ptr := DeZero(original)
		if ptr != nil {
			t.Error("DeZero(0) should return nil")
		}
		got := Zero(ptr)
		if got != 0 {
			t.Errorf("Zero(nil) = %d, want 0", got)
		}
	})
}
