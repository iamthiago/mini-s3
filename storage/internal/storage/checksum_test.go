package storage

import (
	"strings"
	"testing"
)

func TestValueChecksum_Generate(t *testing.T) {
	checksum := NewValueChecksum()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty Input",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Hello World Input",
			input:    "hello world",
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			result, err := checksum.Generate(r)
			if err != nil {
				t.Fatalf("Failed to generate checksum: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected checksum %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValueChecksum_Verify(t *testing.T) {
	checksum := NewValueChecksum()

	tests := []struct {
		name     string
		input    string
		expected string
		want     bool
	}{
		{
			name:     "Valid Checksum",
			input:    "hello world",
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			want:     true,
		},
		{
			name:     "Invalid Checksum",
			input:    "hello world",
			expected: "invalid_checksum",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			result, err := checksum.Verify(r, tt.expected)
			if err != nil {
				t.Fatalf("Failed to verify checksum: %v", err)
			}
			if result != tt.want {
				t.Errorf("Expected checksum %t, got %t", tt.want, result)
			}
		})
	}
}
