package server

import (
	"context"

	"github.com/samwestmoreland/transaction-store/internal/model"
)

type Store interface {
	InsertTransaction(ctx context.Context, tx *model.Transaction) error
	Close()
}
