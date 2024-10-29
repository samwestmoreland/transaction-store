package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"

	"github.com/samwestmoreland/transaction-store/internal/database/mock"
)

func TestServer_HandleTransactionCreate(t *testing.T) {
	tests := []struct {
		name           string
		request        TransactionRequest
		setupMock      func(*mock.MockDB)
		expectedStatus int
		validateDB     func(*testing.T, *mock.MockDB)
	}{
		{
			name: "successful transaction creation",
			request: TransactionRequest{
				TransactionID: uuid.New().String(),
				Amount:        "100.50",
				Timestamp:     time.Now().UTC(),
			},
			setupMock:      func(mock *mock.MockDB) {},
			expectedStatus: http.StatusCreated,
			validateDB: func(t *testing.T, mock *mock.MockDB) {
				txs := mock.GetTransactions()
				if len(txs) != 1 {
					t.Errorf("expected 1 transaction, got %d", len(txs))
				}
			},
		},
		{
			name: "invalid transaction ID",
			request: TransactionRequest{
				TransactionID: "not-a-uuid",
				Amount:        "100.50",
				Timestamp:     time.Now().UTC(),
			},
			setupMock:      func(mock *mock.MockDB) {},
			expectedStatus: http.StatusBadRequest,
			validateDB: func(t *testing.T, mock *mock.MockDB) {
				txs := mock.GetTransactions()
				if len(txs) != 0 {
					t.Errorf("expected 0 transactions, got %d", len(txs))
				}
			},
		},
		{
			name: "invalid amount",
			request: TransactionRequest{
				TransactionID: uuid.New().String(),
				Amount:        "not-a-number",
				Timestamp:     time.Now().UTC(),
			},
			setupMock:      func(mock *mock.MockDB) {},
			expectedStatus: http.StatusBadRequest,
			validateDB: func(t *testing.T, mock *mock.MockDB) {
				txs := mock.GetTransactions()
				if len(txs) != 0 {
					t.Errorf("expected 0 transactions, got %d", len(txs))
				}
			},
		},
		{
			name: "database error",
			request: TransactionRequest{
				TransactionID: uuid.New().String(),
				Amount:        "100.50",
				Timestamp:     time.Now().UTC(),
			},
			setupMock: func(mock *mock.MockDB) {
				mock.SetInsertError(context.DeadlineExceeded)
			},
			expectedStatus: http.StatusInternalServerError,
			validateDB: func(t *testing.T, mock *mock.MockDB) {
				txs := mock.GetTransactions()
				if len(txs) != 0 {
					t.Errorf("expected 0 transactions, got %d", len(txs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			mockDB := mock.NewMockDB()
			tt.setupMock(mockDB)

			server := NewTesting(mockDB, logger)

			body, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/transaction/", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			server.Routes().ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			tt.validateDB(t, mockDB)
		})
	}
}

func TestServer_HandleHealth(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mock.MockDB)
		expectedStatus int
	}{
		{
			name:           "healthy service",
			setupMock:      func(mock *mock.MockDB) {},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unhealthy service",
			setupMock: func(mock *mock.MockDB) {
				mock.SetPingError(context.DeadlineExceeded)
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "closed database",
			setupMock: func(mock *mock.MockDB) {
				mock.Close()
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			mockDB := mock.NewMockDB()
			tt.setupMock(mockDB)

			server := NewTesting(mockDB, logger)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()

			server.Routes().ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestServer_HandleTransactionCreate_InvalidMethod(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockDB := mock.NewMockDB()
	server := NewTesting(mockDB, logger)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/transaction/", nil)
			rec := httptest.NewRecorder()

			server.Routes().ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d for method %s, got %d",
					http.StatusMethodNotAllowed, method, rec.Code)
			}
		})
	}
}

func TestServer_HandleTransactionCreate_InvalidJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockDB := mock.NewMockDB()
	server := NewTesting(mockDB, logger)

	invalidJSON := []byte(`{"transactionId": "invalid json`)

	req := httptest.NewRequest(http.MethodPost, "/api/transaction/", bytes.NewReader(invalidJSON))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid JSON, got %d",
			http.StatusBadRequest, rec.Code)
	}

	txs := mockDB.GetTransactions()
	if len(txs) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(txs))
	}
}
