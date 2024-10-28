package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/samwestmoreland/transaction-store/internal/model"
)

type DB struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func New(ctx context.Context, connString string, logger *zap.Logger) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	db := &DB{
		pool:   pool,
		logger: logger,
	}

	if err := db.ensureSchema(ctx); err != nil {
		pool.Close()

		return nil, err
	}

	return db, nil
}

func (db *DB) InsertTransaction(ctx context.Context, tx *model.Transaction) error {
	_, err := db.pool.Exec(ctx, "INSERT INTO transactions (id, amount, timestamp) VALUES ($1, $2, $3)", tx.ID, tx.Amount, tx.Timestamp)

	return err
}

func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) ensureSchema(ctx context.Context) error {
	const createTable = `
		CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY,
			amount DECIMAL NOT NULL,
			timestamp TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`

	_, err := db.pool.Exec(ctx, createTable)
	if err != nil {
		return err
	}

	return nil
}
