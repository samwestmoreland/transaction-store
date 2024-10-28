package server

import (
	"net/http"

	"github.com/samwestmoreland/transaction-store/internal/model"
	"go.uber.org/zap"
)

type Server struct {
	db     Store
	logger *zap.Logger
}

func New(db Store) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) handleTransactionCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tx := &model.Transaction{}

		if err := s.db.InsertTransaction(r.Context(), tx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
