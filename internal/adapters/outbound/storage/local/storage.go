package local

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
)

const localPrefix = "local://"

type Storage struct {
	baseDir string
}

func NewStorage(baseDir string) *Storage {
	return &Storage{baseDir: baseDir}
}

func (s *Storage) Save(_ context.Context, objectPath, contentType string, data io.Reader) (ports.StoredObject, error) {
	if strings.TrimSpace(objectPath) == "" {
		return ports.StoredObject{}, fmt.Errorf("object path is required")
	}

	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(objectPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return ports.StoredObject{}, fmt.Errorf("local storage mkdir: %w", err)
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return ports.StoredObject{}, fmt.Errorf("local storage create: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, data)
	if err != nil {
		return ports.StoredObject{}, fmt.Errorf("local storage write: %w", err)
	}

	return ports.StoredObject{
		Path:        localPrefix + filepath.ToSlash(objectPath),
		ContentType: contentType,
		SizeBytes:   written,
	}, nil
}

func (s *Storage) Open(_ context.Context, storedPath string) (io.ReadCloser, string, error) {
	resolvedPath := storedPath
	if strings.HasPrefix(storedPath, localPrefix) {
		relative := strings.TrimPrefix(storedPath, localPrefix)
		resolvedPath = filepath.Join(s.baseDir, filepath.FromSlash(relative))
	}

	file, err := os.Open(resolvedPath)
	if err != nil {
		return nil, "", fmt.Errorf("local storage open: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(resolvedPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return file, contentType, nil
}
