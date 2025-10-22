package storage

import (
	"io"
	"time"
)

type ObjectInfo struct {
	Bucket    string
	Object    string
	Size      int64
	Checksum  string
	CreatedAt time.Time
	Path      string
}

type Storage interface {
	Save(bucket, object string, r io.Reader) (*ObjectInfo, error)
	Get(bucket, object string) (io.ReadCloser, *ObjectInfo, error)
	Delete(bucket, object string) error
	Exists(bucket, object string) (bool, error)
	ListObjects(bucket string) ([]*ObjectInfo, error)
}

type LocalStorage struct {
	path string
}

func NewLocalStorage(path string) (*LocalStorage, error) {
	return &LocalStorage{path: path}, nil
}

func (l *LocalStorage) Save(bucket, object string, r io.Reader) (*ObjectInfo, error) {
	return nil, nil
}

func (l *LocalStorage) Get(bucket, object string) (io.ReadCloser, *ObjectInfo, error) {
	return nil, nil, nil
}

func (l *LocalStorage) Delete(bucket, object string) error {
	return nil
}

func (l *LocalStorage) Exists(bucket, object string) (bool, error) {
	return false, nil
}

func (l *LocalStorage) ListObjects(bucket string) ([]*ObjectInfo, error) {
	return nil, nil
}
