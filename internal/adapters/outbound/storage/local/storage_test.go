package local

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestStorage_SaveAndOpen_RoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := NewStorage(dir)

	obj, err := s.Save(context.Background(), "reports/week1.txt", "text/plain", strings.NewReader("hola"))
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if obj.Path != "local://reports/week1.txt" {
		t.Fatalf("unexpected path: %s", obj.Path)
	}
	if obj.SizeBytes != 4 {
		t.Fatalf("unexpected size: %d", obj.SizeBytes)
	}

	r, contentType, err := s.Open(context.Background(), obj.Path)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "hola" {
		t.Fatalf("unexpected content: %s", string(data))
	}
	if contentType != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", contentType)
	}
}

func TestStorage_Save_EmptyPath(t *testing.T) {
	t.Parallel()

	s := NewStorage(t.TempDir())
	_, err := s.Save(context.Background(), "  ", "text/plain", strings.NewReader("x"))
	if err == nil || !strings.Contains(err.Error(), "object path is required") {
		t.Fatalf("expected object path validation error, got: %v", err)
	}
}

func TestStorage_Open_UnknownExtension_DefaultsOctetStream(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := NewStorage(dir)

	obj, err := s.Save(context.Background(), "binary/report.binx", "application/x-custom", strings.NewReader("123"))
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	r, contentType, err := s.Open(context.Background(), obj.Path)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer r.Close()

	if contentType != "application/octet-stream" {
		t.Fatalf("expected default content type, got: %s", contentType)
	}
}
