package bootstrap

import (
	"context"
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/config"
)

func TestNoOpClose_DoesNotPanic(t *testing.T) {
	noOpClose()
}

func TestInitStorage_DefaultReturnsLocalAndNoOpCloser(t *testing.T) {
	cfg := config.Config{
		StorageProvider: "local",
		StorageLocalDir: t.TempDir(),
	}

	storage, closeFn, err := initStorage(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if storage == nil {
		t.Fatal("expected non-nil storage")
	}
	if closeFn == nil {
		t.Fatal("expected non-nil close function")
	}

	// local storage does not hold external connections; close must be safe no-op.
	closeFn()
}

func TestInitStorage_GCSRequiresBucket(t *testing.T) {
	cfg := config.Config{StorageProvider: "gcs"}

	storage, closeFn, err := initStorage(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error when GCS bucket is missing")
	}
	if storage != nil {
		t.Fatal("expected nil storage when bucket is missing")
	}
	if closeFn == nil {
		t.Fatal("expected non-nil close function")
	}

	closeFn()
}
