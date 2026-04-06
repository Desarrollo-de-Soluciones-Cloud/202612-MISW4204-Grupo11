package application

import (
	"context"
	"errors"
	"fmt"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Readiness struct {
	DB Pinger
}

func (r *Readiness) Check(ctx context.Context) error {
	if r == nil || r.DB == nil {
		return errors.New("base de datos no configurada")
	}
	if err := r.DB.Ping(ctx); err != nil {
		return fmt.Errorf("no responde la base de datos: %w", err)
	}
	return nil
}
