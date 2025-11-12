package cmd

import (
	"io"

	"github.com/iamthiago/mini-s3/internal/storage"
)

type mockStorageForTesting struct {
	saveFunc        func(bucket, object string, reader io.Reader) (*storage.ObjectInfo, error)
	listObjectsFunc func(bucket string) ([]*storage.ObjectInfo, error)
	getFunc         func(bucket, object string) (io.ReadCloser, *storage.ObjectInfo, error)
	deleteFunc      func(bucket, object string) error
	existsFunc      func(bucket, object string) (bool, error)
}

func (m *mockStorageForTesting) Save(bucket, object string, reader io.Reader) (*storage.ObjectInfo, error) {
	if m.saveFunc != nil {
		return m.saveFunc(bucket, object, reader)
	}
	return &storage.ObjectInfo{Checksum: "mock-checksum"}, nil
}

func (m *mockStorageForTesting) ListObjects(bucket string) ([]*storage.ObjectInfo, error) {
	if m.listObjectsFunc != nil {
		return m.listObjectsFunc(bucket)
	}
	return []*storage.ObjectInfo{}, nil
}

func (m *mockStorageForTesting) Get(bucket, object string) (io.ReadCloser, *storage.ObjectInfo, error) {
	if m.getFunc != nil {
		return m.getFunc(bucket, object)
	}
	return nil, nil, nil
}

func (m *mockStorageForTesting) Delete(bucket, object string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(bucket, object)
	}
	return nil
}

func (m *mockStorageForTesting) Exists(bucket, object string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(bucket, object)
	}
	return false, nil
}

// withMockStorage temporarily replaces the global storageInstance with a mock
// for testing purposes. Returns a cleanup function that must be called to restore.
func withMockStorage(mock storage.Storage) func() {
	original := storageInstance
	storageInstance = mock
	return func() {
		storageInstance = original
	}
}
