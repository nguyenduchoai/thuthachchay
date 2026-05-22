// Package db dựng pgxpool và helper transaction.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/buocvang/api/internal/config"
)

type Pool = pgxpool.Pool

// New mở pool theo DATABASE_URL. Áp dụng max conns + lifetime.
func New(ctx context.Context, cfg *config.Config) (*Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}
	if cfg.DatabaseMaxOpenConns > 0 {
		pcfg.MaxConns = int32(cfg.DatabaseMaxOpenConns)
	}
	if cfg.DatabaseMaxIdleConns > 0 {
		pcfg.MinConns = int32(cfg.DatabaseMaxIdleConns)
	}
	if cfg.DatabaseConnMaxLifetime > 0 {
		pcfg.MaxConnLifetime = cfg.DatabaseConnMaxLifetime
	} else {
		pcfg.MaxConnLifetime = 30 * time.Minute
	}
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Info().Int32("max_conns", pcfg.MaxConns).Msg("db pool ready")
	return pool, nil
}

// InTx chạy fn trong 1 transaction. Rollback nếu fn lỗi.
func InTx(ctx context.Context, pool *Pool, fn func(pgx.Tx) error) (err error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Warn().Err(rbErr).Msg("tx rollback")
			}
			return
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit tx: %w", commitErr)
		}
	}()
	return fn(tx)
}
