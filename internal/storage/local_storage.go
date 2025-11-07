package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	Get(bucket, object string, expectedChecksum string) (io.ReadCloser, *ObjectInfo, error)
	Delete(bucket, object string) error
	Exists(bucket, object string) (bool, error)
	ListObjects(bucket string) ([]*ObjectInfo, error)
}

type LocalStorage struct {
	path     string
	checksum Checksum
}

func NewLocalStorage(path string, checkSum Checksum) *LocalStorage {
	return &LocalStorage{path: path, checksum: checkSum}
}

func (l *LocalStorage) Save(bucket, object string, r io.Reader) (*ObjectInfo, error) {
	createdAt := time.Now()

	// Create bucket directory
	path := filepath.Join(l.path, bucket)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	// Create file
	filePath := filepath.Join(path, object)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Split the stream with a pipe
	pr, pw := io.Pipe()
	defer pr.Close()

	// Compute checksum in a goroutine
	checksumCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		checksum, err := l.checksum.Generate(pr)
		if err != nil {
			errCh <- err
			return
		}
		checksumCh <- checksum
	}()

	// Write to file and pipe it
	teeReader := io.TeeReader(r, pw)
	size, err := io.Copy(file, teeReader)
	pw.Close()

	if err != nil {
		return nil, err
	}

	// Wait for checksum to be computed
	var checksum string
	select {
	case checksum = <-checksumCh:
	case err := <-errCh:
		return nil, err
	}

	newObj := &ObjectInfo{
		Bucket:    bucket,
		Object:    object,
		Size:      size,
		Checksum:  checksum,
		CreatedAt: createdAt,
		Path:      filePath,
	}

	return newObj, nil
}

func (l *LocalStorage) Get(bucket, object string, expectedChecksum string) (io.ReadCloser, *ObjectInfo, error) {
	filepath := filepath.Join(l.path, bucket, object)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	isValidChecksum, calculatedChecksum, err := l.checksum.Verify(file, expectedChecksum)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	if !isValidChecksum {
		file.Close()
		return nil, nil, &ErrInvalidChecksum{
			Got:      calculatedChecksum,
			Expected: expectedChecksum,
		}
	}

	// Reset file pointer to the beginning
	// as it was consumed by the checksum verification
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	objInfo := &ObjectInfo{
		Bucket:    bucket,
		Object:    object,
		Size:      info.Size(),
		Checksum:  calculatedChecksum,
		CreatedAt: info.ModTime(),
		Path:      filepath,
	}

	return file, objInfo, nil
}

func (l *LocalStorage) Delete(bucket, object string) error {
	filePath := filepath.Join(l.path, bucket, object)
	return os.Remove(filePath)
}

func (l *LocalStorage) Exists(bucket, object string) (bool, error) {
	filePath := filepath.Join(l.path, bucket, object)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LocalStorage) ListObjects(bucket string) ([]*ObjectInfo, error) {
	filePath := filepath.Join(l.path, bucket)
	files, err := os.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	var infos []*ObjectInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(filePath, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return nil, err
		}
		infos = append(infos, &ObjectInfo{
			Bucket:    bucket,
			Object:    file.Name(),
			Size:      fileInfo.Size(),
			Checksum:  "",
			CreatedAt: fileInfo.ModTime(),
			Path:      filePath,
		})
	}

	return infos, nil
}

type ErrInvalidChecksum struct {
	Got      string
	Expected string
}

func (e *ErrInvalidChecksum) Error() string {
	return fmt.Sprintf("invalid checksum: got %s, expected %s", e.Got, e.Expected)
}
