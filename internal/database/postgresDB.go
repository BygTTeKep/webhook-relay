package database

import (
	"context"
	"database/sql"
	"fmt"
	"webhook-relay/internal/config"
)

type PGRepository struct {
	DB *sql.DB
}

func NewPG(cfg *config.DBConfig) (*PGRepository, error) {
	var connStr string

	if cfg.Driver == "postgres" {
		connStr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name,
		)
	}
	db, err := sql.Open(cfg.Driver, connStr)
	if err != nil {
		return nil, err
	}
	return  &PGRepository{DB: db}, nil
}

func (pg *PGRepository) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := pg.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func ()  {
		if p := recover(); p!= nil {
			tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v", rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return  nil
}

func (pg *PGRepository) Close() error {
	return pg.DB.Close()
}