package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/samwestmoreland/transaction-store/internal/model"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}

func (db *DB) InsertTransaction(ctx context.Context, tx *model.Transaction) error {
	_, err := db.pool.Exec(ctx, "INSERT INTO transactions (id, amount, timestamp) VALUES ($1, $2, $3)", tx.ID, tx.Amount, tx.Timestamp)

	return err
}

func (db *DB) Close() {
	db.pool.Close()
}
