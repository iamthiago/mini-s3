package storage

import (
	"bytes"
	"errors"
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
	tempDir, err := os.MkdirTemp("", "storage-exists-test")
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

func TestLocalStorage_Get(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-get-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	checksum := NewValueChecksum()
	storage := NewLocalStorage(tempDir, checksum)

	t.Run("Returns error when file or bucket does not exist", func(t *testing.T) {
		_, _, err := storage.Get("invalid-bucket", "invalid-file.txt", "invalid-checksum")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Returns error when file cannot be read", func(t *testing.T) {
		_, err := storage.Save("test-bucket", "test-file.txt", &errorReader{err: io.ErrUnexpectedEOF})
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Returns error when checksum does not match", func(t *testing.T) {
		bucket := "test-bucket"
		fileName := "test-file.txt"

		_, err := storage.Save(bucket, fileName, strings.NewReader("Hello World!"))
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		_, _, err = storage.Get(bucket, fileName, "invalid-checksum")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		var errInvalidChecksum *ErrInvalidChecksum
		if !errors.As(err, &errInvalidChecksum) {
			t.Errorf("Expected ErrInvalidChecksum, got %v", err)
		}

		if errInvalidChecksum.Got == "" {
			t.Errorf("Expected Got field to be set")
		}

		if errInvalidChecksum.Expected != "invalid-checksum" {
			t.Errorf("Expected Expected field to be 'invalid-checksum', got '%s'", errInvalidChecksum.Expected)
		}

		// Test Error() method
		errorMsg := errInvalidChecksum.Error()
		if errorMsg == "" {
			t.Errorf("Expected error message to be non-empty")
		}
		if !strings.Contains(errorMsg, "invalid checksum") {
			t.Errorf("Expected error message to contain 'invalid checksum', got '%s'", errorMsg)
		}
	})

	t.Run("Returns error when checksum verification fails", func(t *testing.T) {
		bucket := "test-bucket"
		fileName := "test-checksum-error.txt"

		_, err := storage.Save(bucket, fileName, strings.NewReader("Hello World!"))
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		// Create a storage with error-prone checksum
		errorChecksum := &errorChecksum{}
		errorStorage := NewLocalStorage(tempDir, errorChecksum)

		_, _, err = errorStorage.Get(bucket, fileName, "any-checksum")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Successfully retrieves existing file", func(t *testing.T) {
		text := "Hello World!"
		bucket := "test-bucket"
		fileName := "test-get-file.txt"

		reader := strings.NewReader(text)
		savedInfo, err := storage.Save(bucket, fileName, reader)
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		// Use the checksum from the saved file
		file, objInfo, err := storage.Get(bucket, fileName, savedInfo.Checksum)
		if err != nil {
			t.Fatalf("Failed to get file: %v", err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if string(content) != text {
			t.Errorf("File content does not match. Expected '%s', got '%s'", text, content)
		}

		if objInfo.Bucket != bucket {
			t.Errorf("Expected bucket to be '%s', got '%s'", bucket, objInfo.Bucket)
		}

		if objInfo.Object != fileName {
			t.Errorf("Expected object name to be %s, got %s", fileName, objInfo.Object)
		}

		if objInfo.Checksum != savedInfo.Checksum {
			t.Errorf("Expected checksum to be '%s', got '%s'", savedInfo.Checksum, objInfo.Checksum)
		}
	})
}

func TestLocalStorage_Delete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-delete-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	checksum := NewValueChecksum()
	storage := NewLocalStorage(tempDir, checksum)

	t.Run("Returns error when file or bucket does not exist", func(t *testing.T) {
		err := storage.Delete("invalid-bucket", "invalid-file.txt")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Successfully deletes existing file", func(t *testing.T) {
		info, err := storage.Save("test-bucket", "test-delete-file.txt", strings.NewReader("Hello World!"))
		if err != nil {
			t.Fatalf("Failed to save file: %v", err)
		}

		err = storage.Delete(info.Bucket, info.Object)
		if err != nil {
			t.Fatalf("Failed to delete file: %v", err)
		}
	})
}

func TestLocalStorage_ListObjects(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-list-objects-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	checksum := NewValueChecksum()
	storage := NewLocalStorage(tempDir, checksum)

	t.Run("Returns error when bucket does not exist", func(t *testing.T) {
		_, err := storage.ListObjects("invalid-bucket")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Returns empty list when bucket is empty", func(t *testing.T) {
		bucketPath := filepath.Join(tempDir, "empty-bucket")
		err := os.Mkdir(bucketPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create empty bucket: %v", err)
		}

		objects, err := storage.ListObjects("empty-bucket")
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		if len(objects) != 0 {
			t.Errorf("Expected empty list, got %d objects", len(objects))
		}
	})

	t.Run("Returns list of objects in bucket", func(t *testing.T) {
		bucket := "test-bucket"
		objects := []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"}

		for _, object := range objects {
			_, err := storage.Save(bucket, object, strings.NewReader("Hello World!"))
			if err != nil {
				t.Fatalf("Failed to save file: %v", err)
			}
		}

		listedObjects, err := storage.ListObjects(bucket)
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		if len(listedObjects) != len(objects) {
			t.Errorf("Expected %d objects, got %d", len(objects), len(listedObjects))
		}
	})
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

// errorChecksum is a mock that always return an error
type errorChecksum struct{}

func (e *errorChecksum) Generate(r io.Reader) (string, error) {
	return "", errors.New("checksum generation error")
}

func (e *errorChecksum) Verify(r io.Reader, expected string) (bool, string, error) {
	return false, "", errors.New("checksum verification error")
}
