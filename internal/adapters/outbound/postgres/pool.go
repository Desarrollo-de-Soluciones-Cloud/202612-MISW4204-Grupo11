package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	inner *pgxpool.Pool
}

func NewPool(ctx context.Context, url string) (*Pool, func(), error) {
	if url == "" {
		return nil, nil, fmt.Errorf("la URL de la base de datos está vacía")
	}
	p, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, nil, err
	}
	return &Pool{inner: p}, func() { p.Close() }, nil
}

func (p *Pool) Ping(ctx context.Context) error {
	if p == nil || p.inner == nil {
		return fmt.Errorf("pool no inicializado")
	}
	return p.inner.Ping(ctx)
}
