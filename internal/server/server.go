package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samwestmoreland/transaction-store/internal/model"
	"go.uber.org/zap"
)

type Server struct {
	db      Store
	logger  *zap.Logger
	metrics *metrics
}

// TransactionRequest represents the incoming JSON payload
type TransactionRequest struct {
	TransactionID string    `json:"transactionId"`
	Amount        string    `json:"amount"` // Using string to match the example curl command
	Timestamp     time.Time `json:"timestamp"`
}

func New(db Store, logger *zap.Logger) *Server {
	return &Server{
		db:      db,
		logger:  logger,
		metrics: newMetrics(),
	}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/transaction/", s.handleTransactionCreate())
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", s.handleHealth())

	return mux
}

func (s *Server) handleTransactionCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			s.logger.Warn("invalid method",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Start timing the request
		start := time.Now()
		defer func() {
			s.metrics.requestDuration.Observe(time.Since(start).Seconds())
		}()

		// Parse request body
		var req TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.logger.Warn("failed to decode request",
				zap.Error(err))
			http.Error(w, "invalid request body", http.StatusBadRequest)
			s.metrics.requestErrors.Inc()
			return
		}

		s.logger.Debug("successfully decoded request",
			zap.String("transaction_id", req.TransactionID),
			zap.String("amount", req.Amount),
			zap.Time("timestamp", req.Timestamp))

		// Validate and parse the transaction ID
		txID, err := uuid.Parse(req.TransactionID)
		if err != nil {
			s.logger.Warn("invalid transaction ID",
				zap.String("id", req.TransactionID),
				zap.Error(err))
			http.Error(w, "invalid transaction ID", http.StatusBadRequest)
			s.metrics.requestErrors.Inc()
			return
		}

		// Parse the amount
		amount, err := strconv.ParseFloat(req.Amount, 64)
		if err != nil {
			s.logger.Warn("invalid amount",
				zap.String("amount", req.Amount),
				zap.Error(err))
			http.Error(w, "invalid amount", http.StatusBadRequest)
			s.metrics.requestErrors.Inc()
			return
		}

		// Create the transaction
		tx := &model.Transaction{
			ID:        txID,
			Amount:    amount,
			Timestamp: req.Timestamp,
		}

		// Insert into database
		if err := s.db.InsertTransaction(r.Context(), tx); err != nil {
			s.logger.Error("failed to insert transaction",
				zap.Error(err),
				zap.String("txID", tx.ID.String()))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			s.metrics.requestErrors.Inc()
			return
		}

		// Increment success counter
		s.metrics.requestSuccess.Inc()

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
			"id":     tx.ID.String(),
		})
	}
}

func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		if err := s.db.Ping(ctx); err != nil {
			s.logger.Error("health check failed", zap.Error(err))
			http.Error(w, "unhealthy", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}
}
