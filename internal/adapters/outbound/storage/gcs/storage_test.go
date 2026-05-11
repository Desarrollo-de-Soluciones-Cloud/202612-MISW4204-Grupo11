package gcs

import (
	"context"
	"strings"
	"testing"
)

func TestStorage_Save_EmptyPath(t *testing.T) {
	t.Parallel()

	s := &Storage{bucket: "test-bucket"}
	_, err := s.Save(context.Background(), "", "text/plain", strings.NewReader("x"))
	if err == nil || !strings.Contains(err.Error(), "object path is required") {
		t.Fatalf("expected object path validation error, got: %v", err)
	}
}

func TestStorage_Close_NilSafe(t *testing.T) {
	t.Parallel()

	var s *Storage
	if err := s.Close(); err != nil {
		t.Fatalf("expected nil close error, got: %v", err)
	}

	s = &Storage{}
	if err := s.Close(); err != nil {
		t.Fatalf("expected nil close error with nil client, got: %v", err)
	}
}
