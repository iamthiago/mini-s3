package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

type Checksum interface {
	Generate(r io.Reader) (string, error)
	Verify(r io.Reader, expected string) (bool, error)
}

type ValueChecksum struct {
}

func NewValueChecksum() (*ValueChecksum, error) {
	return &ValueChecksum{}, nil
}

// Generate computes a SHA-256 hash of the input stream.
// It returns the checksum as a hex-encoded string and the number of bytes read.
func (v *ValueChecksum) Generate(r io.Reader) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, r)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Verify reads from r and compares its checksum to the provided expected value.
// Returns true if the checksums match.
func (v *ValueChecksum) Verify(r io.Reader, expected string) (bool, error) {
	calculated, err := v.Generate(r)
	if err != nil {
		return false, err
	}
	return calculated == expected, nil
}
