package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	baseDir string
}

// NewLocalStorage initializes a local filesystem storage backend.
func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir}
}

// Store saves the file data to disk under baseDir/key and returns the storage path (key).
func (s *LocalStorage) Store(ctx context.Context, key string, contentType string, data io.Reader) (string, error) {
	fullPath := filepath.Join(s.baseDir, key)

	// Automatically create required subdirectories
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create storage file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("failed to write storage file data: %w", err)
	}

	return key, nil
}

// Open returns an io.ReadCloser for the file at path (relative to baseDir).
func (s *LocalStorage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, path)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("storage object not found: %w", err)
		}
		return nil, fmt.Errorf("failed to open storage file: %w", err)
	}
	return file, nil
}

// Delete removes the file at storagePath (relative to baseDir).
func (s *LocalStorage) Delete(ctx context.Context, storagePath string) error {
	fullPath := filepath.Join(s.baseDir, storagePath)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete storage file: %w", err)
	}
	return nil
}
