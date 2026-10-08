package database

import (
	"context"
	"fmt"
	"time"
	"webhook-relay/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGRepository struct {
	DB *pgxpool.Pool
}

func NewPG(ctx context.Context, cfg *config.DBConfig) (*PGRepository, error) {
	conf, err := pgxpool.ParseConfig(cfg.Dsn)
	if err != nil {
		return nil, err
	}
	conf.MaxConns = 20
	conf.MinConns = 4
	conf.MaxConnLifetime = time.Hour
	conf.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, err
	}
	// if err := pool.Ping(ctx); err != nil {
	// 	return nil, err
	// }
	return &PGRepository{
		DB: pool,
	}, nil
}

func (pg *PGRepository) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := pg.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback failed: %v", rbErr)
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (pg *PGRepository) Close() {
	pg.DB.Close()
}
