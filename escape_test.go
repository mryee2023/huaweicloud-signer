package huaweicloud

import (
	"testing"
)

func TestShouldEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected bool
	}{
		{"alphanumeric", 'a', false},
		{"special_chars", '-', false},
		{"space", ' ', true},
		{"slash", '/', true},
		{"backslash", '\\', true},
		{"high_byte", 128, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldEscape(tt.input); got != tt.expected {
				t.Errorf("shouldEscape(%c) = %v, expected %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty_string", "", ""},
		{"simple_text", "hello", "hello"},
		{"alphanumeric", "abc123", "abc123"},
		{"allowed_special_chars", "abc-_.~", "abc-_.~"},
		{"space", " ", "%20"},
		{"url_like", "http://example.com", "http%3A%2F%2Fexample.com"},
		{"query_string", "a=b&c=d", "a%3Db%26c%3Dd"},

		{"special_chars", "!@#$%^&*()", "%21%40%23%24%25%5E%26%2A%28%29"},
		{"high_byte", string([]byte{128}), "%80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escape(tt.input); got != tt.expected {
				t.Errorf("escape(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEscapeEmptyString(t *testing.T) {
	if got := escape(""); got != "" {
		t.Errorf("escape(\"\") = %q, expected \"\"", got)
	}
}

func TestEscapeNoEscapeNeeded(t *testing.T) {
	input := "abc-_.~123"
	if got := escape(input); got != input {
		t.Errorf("escape(%q) = %q, expected %q", input, got, input)
	}
}

func TestEscapeHexCount(t *testing.T) {
	input := "!@#$%^&*()"
	expectedCount := 10 // 每个特殊字符都需要转义
	hexCount := 0
	for i := 0; i < len(input); i++ {
		if shouldEscape(input[i]) {
			hexCount++
		}
	}
	if hexCount != expectedCount {
		t.Errorf("hexCount = %d, expected %d", hexCount, expectedCount)
	}
}
