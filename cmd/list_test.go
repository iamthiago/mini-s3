package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/iamthiago/mini-s3/internal/storage"
)

func TestListCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupStorage   func() *mockStorageForTesting
		wantErr        bool
		expectedOutput string
	}{
		{
			name: "successful list with objects",
			args: []string{"test-bucket"},
			setupStorage: func() *mockStorageForTesting {
				return &mockStorageForTesting{
					listObjectsFunc: func(bucket string) ([]*storage.ObjectInfo, error) {
						if bucket != "test-bucket" {
							t.Errorf("expected bucket 'test-bucket', got '%s'", bucket)
						}
						return []*storage.ObjectInfo{
							{
								Bucket:    "test-bucket",
								Object:    "file1.txt",
								Size:      1024,
								CreatedAt: time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
								Checksum:  "checksum1",
							},
							{
								Bucket:    "test-bucket",
								Object:    "file2.txt",
								Size:      5242880, // 5 MB
								CreatedAt: time.Date(2024, 1, 15, 15, 30, 45, 0, time.UTC),
								Checksum:  "checksum2",
							},
						}, nil
					},
				}
			},
			wantErr:        false,
			expectedOutput: "file1.txt",
		},
		{
			name: "empty bucket",
			args: []string{"test-bucket"},
			setupStorage: func() *mockStorageForTesting {
				return &mockStorageForTesting{
					listObjectsFunc: func(bucket string) ([]*storage.ObjectInfo, error) {
						return []*storage.ObjectInfo{}, nil
					},
				}
			},
			wantErr:        false,
			expectedOutput: "No objects found",
		},
		{
			name:           "missing bucket argument",
			args:           []string{},
			setupStorage:   func() *mockStorageForTesting { return &mockStorageForTesting{} },
			wantErr:        false,
			expectedOutput: "Usage: mini-s3 list <bucket-name>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock storage
			cleanup := withMockStorage(tt.setupStorage())
			defer cleanup()

			// Capture output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute command
			listCmd.Run(listCmd, tt.args)

			// Restore stdout and read output
			_ = w.Close()
			os.Stdout = old
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			// Check output contains expected string
			if !bytes.Contains([]byte(output), []byte(tt.expectedOutput)) {
				t.Errorf("expected output to contain '%s', got '%s'", tt.expectedOutput, output)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    100,
			expected: "100 B",
		},
		{
			name:     "kilobytes",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "megabytes",
			bytes:    1048576,
			expected: "1.0 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1073741824,
			expected: "1.0 GB",
		},
		{
			name:     "fractional KB",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestListCommandHasRequiredFields(t *testing.T) {
	if listCmd.Use == "" {
		t.Error("listCmd.Use should not be empty")
	}
	if listCmd.Short == "" {
		t.Error("listCmd.Short should not be empty")
	}
}
