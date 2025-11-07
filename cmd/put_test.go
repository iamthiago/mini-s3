package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/iamthiago/mini-s3/internal/storage"
)

func TestPutCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name           string
		args           []string
		setupStorage   func() *mockStorageForTesting
		wantErr        bool
		expectedOutput string
	}{
		{
			name: "successful put",
			args: []string{"test-bucket", testFile},
			setupStorage: func() *mockStorageForTesting {
				return &mockStorageForTesting{
					saveFunc: func(bucket, object string, reader io.Reader) (*storage.ObjectInfo, error) {
						if bucket != "test-bucket" {
							t.Errorf("expected bucket 'test-bucket', got '%s'", bucket)
						}
						if object != "test.txt" {
							t.Errorf("expected object 'test.txt', got '%s'", object)
						}
						// Verify content
						content, _ := io.ReadAll(reader)
						if !bytes.Equal(content, testContent) {
							t.Errorf("expected content '%s', got '%s'", testContent, content)
						}
						return &storage.ObjectInfo{Checksum: "checksum123"}, nil
					},
				}
			},
			wantErr:        false,
			expectedOutput: "Successfully added test.txt to bucket test-bucket",
		},
		{
			name:           "missing arguments",
			args:           []string{"test-bucket"},
			setupStorage:   func() *mockStorageForTesting { return &mockStorageForTesting{} },
			wantErr:        false,
			expectedOutput: "Usage: mini-s3 put <bucket-name> <object-name>",
		},
		{
			name:           "no arguments",
			args:           []string{},
			setupStorage:   func() *mockStorageForTesting { return &mockStorageForTesting{} },
			wantErr:        false,
			expectedOutput: "Usage: mini-s3 put <bucket-name> <object-name>",
		},
		{
			name: "file does not exist",
			args: []string{"test-bucket", "/nonexistent/file.txt"},
			setupStorage: func() *mockStorageForTesting {
				return &mockStorageForTesting{}
			},
			wantErr:        false,
			expectedOutput: "Failed to open file:",
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
			putCmd.Run(putCmd, tt.args)

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

func TestPutCommandHasRequiredFields(t *testing.T) {
	if putCmd.Use == "" {
		t.Error("putCmd.Use should not be empty")
	}
	if putCmd.Short == "" {
		t.Error("putCmd.Short should not be empty")
	}
}
