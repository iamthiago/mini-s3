package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/iamthiago/mini-s3/internal/storage"
)

func TestGetCommand(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("Failed to create dest directory: %v", err)
	}

	testFile := filepath.Join(sourceDir, "test.txt")
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
		verifyFile     bool
	}{
		{
			name: "get file successfully",
			args: []string{"test-bucket", "test.txt", destDir},
			setupStorage: func() *mockStorageForTesting {
				return &mockStorageForTesting{
					getFunc: func(bucket, object string) (io.ReadCloser, *storage.ObjectInfo, error) {
						file, err := os.Open(testFile)
						if err != nil {
							return nil, nil, err
						}
						return file, &storage.ObjectInfo{Object: object}, nil
					},
				}
			},
			wantErr:        false,
			expectedOutput: "Successfully saved test.txt to " + destDir,
			verifyFile:     true,
		},
		{
			name:           "missing arguments",
			args:           []string{"test-bucket"},
			setupStorage:   func() *mockStorageForTesting { return &mockStorageForTesting{} },
			wantErr:        false,
			expectedOutput: "Usage: mini-s3 get <bucket-name> <object-name> <output-dir>",
		},
		{
			name:           "no arguments",
			args:           []string{},
			setupStorage:   func() *mockStorageForTesting { return &mockStorageForTesting{} },
			wantErr:        false,
			expectedOutput: "Usage: mini-s3 get <bucket-name> <object-name> <output-dir>",
		},
		{
			name: "file does not exist",
			args: []string{"test-bucket", "nonexistent.txt", destDir},
			setupStorage: func() *mockStorageForTesting {
				return &mockStorageForTesting{
					getFunc: func(bucket, object string) (io.ReadCloser, *storage.ObjectInfo, error) {
						return nil, nil, os.ErrNotExist
					},
				}
			},
			wantErr:        false,
			expectedOutput: "Error getting object: nonexistent.txt. file does not exist",
		},
		{
			name: "error when creating the file",
			args: []string{"test-bucket", "nonexistent.txt", "/nonexistent/dir"},
			setupStorage: func() *mockStorageForTesting {
				return &mockStorageForTesting{
					getFunc: func(bucket, object string) (io.ReadCloser, *storage.ObjectInfo, error) {
						return nil, nil, os.ErrPermission
					},
				}
			},
			wantErr:        true,
			expectedOutput: "Error getting object: nonexistent.txt. permission denied",
		},
		{
			name: "error when writing to the destination file",
			args: []string{"test-bucket", "nonexistent.txt", destDir},
			setupStorage: func() *mockStorageForTesting {
				return &mockStorageForTesting{
					getFunc: func(bucket, object string) (io.ReadCloser, *storage.ObjectInfo, error) {
						return &errorReader{}, &storage.ObjectInfo{Object: object}, nil
					},
				}
			},
			wantErr:        true,
			expectedOutput: "Error writing to file:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := withMockStorage(tt.setupStorage())
			defer cleanup()

			// Capture output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			getCmd.Run(getCmd, tt.args)

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

			// Verify file was written correctly
			if tt.verifyFile {
				downloadedFile := filepath.Join(destDir, "test.txt")
				content, err := os.ReadFile(downloadedFile)
				if err != nil {
					t.Errorf("Failed to read downloaded file: %v", err)
				}
				if !bytes.Equal(content, testContent) {
					t.Errorf("Downloaded file content does not match. Expected '%s', got '%s'", testContent, content)
				}
			}
		})
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (e *errorReader) Close() error {
	return nil
}
