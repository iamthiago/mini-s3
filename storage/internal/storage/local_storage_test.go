package storage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorage_Save(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-save-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	checksum := NewValueChecksum()
	storage := NewLocalStorage(tempDir, checksum)

	t.Run("Saves file successfully", func(t *testing.T) {
		content := "Hello World!"
		reader := strings.NewReader(content)

		fileInfo, err := storage.Save("test-bucket", "test-file.txt", reader)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		// Verify ObjectInfo
		if fileInfo.Bucket != "test-bucket" {
			t.Errorf("Expected bucket to be 'test-bucket', got '%s'", fileInfo.Bucket)
		}

		if fileInfo.Object != "test-file.txt" {
			t.Errorf("Expected object to be 'test-file.txt', got '%s'", fileInfo.Object)
		}

		if fileInfo.Size != int64(len(content)) {
			t.Errorf("Expected size to be %d, got %d", len(content), fileInfo.Size)
		}

		if fileInfo.Checksum == "" {
			t.Errorf("Checksum should not be empty")
		}

		// Verify file exists and has correct content
		savedContent, err := os.ReadFile(fileInfo.Path)
		if err != nil {
			t.Fatalf("Failed to read saved file: %v", err)
		}

		if string(savedContent) != content {
			t.Errorf("File content does not match. Expected '%s', got '%s'", content, savedContent)
		}
	})

	t.Run("Creates directory if it does not exist", func(t *testing.T) {
		reader := strings.NewReader("Hello World!")
		_, err := storage.Save("new-bucket", "new-file.txt", reader)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		bucketPath := filepath.Join(tempDir, "new-bucket")
		if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
			t.Errorf("Bucket does not exist at path: %s", bucketPath)
		}
	})

	t.Run("Handles empty file", func(t *testing.T) {
		reader := strings.NewReader("")
		info, err := storage.Save("empty-bucket", "empty-file.txt", reader)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		if info.Size != 0 {
			t.Errorf("Expected size 0, got %d", info.Size)
		}
	})

	t.Run("Handles large file", func(t *testing.T) {
		// Create 1MB of data
		largeData := bytes.Repeat([]byte("a"), 1024*1024)
		reader := bytes.NewReader(largeData)

		info, err := storage.Save("large-bucket", "large-file.txt", reader)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		if info.Size != int64(len(largeData)) {
			t.Errorf("Expected size %d, got %d", len(largeData), info.Size)
		}
	})

	t.Run("Checksum is correct", func(t *testing.T) {
		content := "Hello World!"
		reader := strings.NewReader(content)

		info, err := storage.Save("checksum-bucket", "checksum-file.txt", reader)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		file, err := os.Open(info.Path)
		if err != nil {
			t.Fatalf("Failed to open file: %v", err)
		}
		defer file.Close()

		expectedChecksum, err := checksum.Generate(file)
		if err != nil {
			t.Fatalf("Failed to generate checksum: %v", err)
		}

		if info.Checksum != expectedChecksum {
			t.Errorf("Expected checksum '%s', got '%s'", expectedChecksum, info.Checksum)
		}
	})

	t.Run("Overwrites existing file", func(t *testing.T) {
		bucket := "overwrite-bucket"
		object := "overwrite-file.txt"

		// Save file for the first time
		reader1 := strings.NewReader("original content")
		_, err := storage.Save(bucket, object, reader1)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		// Now overwrite it with new content
		newContent := "new content"
		reader2 := strings.NewReader(newContent)
		info, err := storage.Save(bucket, object, reader2)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		// Verify new content
		savedContent, err := os.ReadFile(info.Path)
		if err != nil {
			t.Fatalf("Failed to read saved file: %v", err)
		}
		if string(savedContent) != newContent {
			t.Errorf("File content does not match. Expected '%s', got '%s'", newContent, savedContent)
		}
	})

	t.Run("Handles reader error", func(t *testing.T) {
		errorReader := &errorReader{err: io.ErrUnexpectedEOF}
		_, err := storage.Save("error-bucket", "error-file.txt", errorReader)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestLocalStorage_Exists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-exist-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	checksum := NewValueChecksum()
	storage := NewLocalStorage(tempDir, checksum)

	t.Run("Returns false if file or bucket does not exist", func(t *testing.T) {
		fileExists, err := storage.Exists("test-bucket", "non-existing-file.txt")
		if err != nil {
			t.Fatalf("Failed to check if file exists: %v", err)
		}
		if fileExists {
			t.Errorf("Expected file to not exist")
		}
	})

	t.Run("Returns true when file exists in bucket", func(t *testing.T) {
		_, err := storage.Save("test-bucket", "test-file.txt", strings.NewReader("Hello World!"))
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		fileExists, err := storage.Exists("test-bucket", "test-file.txt")
		if err != nil {
			t.Fatalf("Failed to check if file exists: %v", err)
		}
		if !fileExists {
			t.Errorf("Expected file to exist")
		}
	})

	t.Run("Returns error on stat failure", func(t *testing.T) {
		restrictedDir := filepath.Join(tempDir, "restricted")
		err := os.Mkdir(restrictedDir, 0000)
		if err != nil {
			t.Skipf("Failed to create restricted directory: %v", err)
		}
		defer func() {
			err := os.Chmod(restrictedDir, 0755)
			if err != nil {
				t.Fatalf("Failed to change restricted directory permissions: %v", err)
			}
		}()

		_, err = storage.Exists("restricted", "file.txt")
		if err == nil {
			t.Errorf("Expected error when accessing restricted directory")
		}
	})
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}
