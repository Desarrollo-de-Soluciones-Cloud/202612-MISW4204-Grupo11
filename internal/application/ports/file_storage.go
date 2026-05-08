package ports

import (
	"context"
	"io"
)

type StoredObject struct {
	Path        string
	ContentType string
	SizeBytes   int64
}

type FileStorage interface {
	Save(ctx context.Context, objectPath, contentType string, data io.Reader) (StoredObject, error)
	Open(ctx context.Context, storedPath string) (io.ReadCloser, string, error)
}
