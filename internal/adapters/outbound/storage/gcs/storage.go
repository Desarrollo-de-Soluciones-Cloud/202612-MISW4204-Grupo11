package gcs

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
)

const gcsPrefix = "gs://"

type Storage struct {
	client *storage.Client
	bucket string
}

func NewStorage(ctx context.Context, bucket string) (*Storage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs create client: %w", err)
	}
	return &Storage{client: client, bucket: bucket}, nil
}

func (s *Storage) Save(ctx context.Context, objectPath, contentType string, data io.Reader) (ports.StoredObject, error) {
	if strings.TrimSpace(objectPath) == "" {
		return ports.StoredObject{}, fmt.Errorf("object path is required")
	}

	writer := s.client.Bucket(s.bucket).Object(objectPath).NewWriter(ctx)
	writer.ContentType = contentType
	written, err := io.Copy(writer, data)
	if err != nil {
		_ = writer.Close()
		return ports.StoredObject{}, fmt.Errorf("gcs write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return ports.StoredObject{}, fmt.Errorf("gcs close writer: %w", err)
	}

	return ports.StoredObject{
		Path:        gcsPrefix + s.bucket + "/" + objectPath,
		ContentType: contentType,
		SizeBytes:   written,
	}, nil
}

func (s *Storage) Open(ctx context.Context, storedPath string) (io.ReadCloser, string, error) {
	objectPath := storedPath
	if strings.HasPrefix(storedPath, gcsPrefix) {
		trimmed := strings.TrimPrefix(storedPath, gcsPrefix)
		prefix := s.bucket + "/"
		if strings.HasPrefix(trimmed, prefix) {
			objectPath = strings.TrimPrefix(trimmed, prefix)
		}
	}

	reader, err := s.client.Bucket(s.bucket).Object(objectPath).NewReader(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("gcs open object: %w", err)
	}

	contentType := reader.Attrs.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return reader, contentType, nil
}

func (s *Storage) Close() error {
	return s.client.Close()
}
