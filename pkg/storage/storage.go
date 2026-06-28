// Package storage provides pluggable storage abstractions for files.
package storage

import (
	"context"
	"io"
)

// Storage defines the interface for file storage operations.
type Storage interface {
	// Store saves the file data and returns the storage path (key).
	Store(ctx context.Context, key string, contentType string, data io.Reader) (storagePath string, err error)

	// Open returns an io.ReadCloser for the file at path.
	Open(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes the file at storagePath.
	Delete(ctx context.Context, storagePath string) error
}
