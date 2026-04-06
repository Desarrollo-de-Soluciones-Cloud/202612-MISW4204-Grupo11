package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
)

type fakeDB struct {
	err error
}

func (f *fakeDB) Ping(ctx context.Context) error {
	return f.err
}

func TestReadiness_sinDB(t *testing.T) {
	r := &application.Readiness{DB: nil}
	if err := r.Check(context.Background()); err == nil {
		t.Fatal("se esperaba error")
	}
}

func TestReadiness_ok(t *testing.T) {
	r := &application.Readiness{DB: &fakeDB{}}
	if err := r.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestReadiness_pingFalla(t *testing.T) {
	r := &application.Readiness{DB: &fakeDB{err: errors.New("fallo")}}
	if err := r.Check(context.Background()); err == nil {
		t.Fatal("se esperaba error")
	}
}
