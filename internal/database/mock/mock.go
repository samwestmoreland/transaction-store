package mock

import (
	"context"
	"errors"
	"sync"

	"github.com/samwestmoreland/transaction-store/internal/model"
)

var (
	ErrDBClosed = errors.New("database closed")
)

// MockDB implements the Store interface for testing
type MockDB struct {
	transactions []*model.Transaction
	mu           sync.RWMutex
	closed       bool
	// Error states for testing different scenarios
	insertError error
	pingError   error
}

func NewMockDB() *MockDB {
	return &MockDB{
		transactions: make([]*model.Transaction, 0),
	}
}

func (m *MockDB) InsertTransaction(ctx context.Context, tx *model.Transaction) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if m.insertError != nil {
		return m.insertError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrDBClosed
	}

	m.transactions = append(m.transactions, tx)
	return nil
}

func (m *MockDB) Ping(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if m.pingError != nil {
		return m.pingError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrDBClosed
	}

	return nil
}

func (m *MockDB) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
}

func (m *MockDB) SetInsertError(err error) {
	m.insertError = err
}

func (m *MockDB) SetPingError(err error) {
	m.pingError = err
}

func (m *MockDB) GetTransactions() []*model.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*model.Transaction, len(m.transactions))
	copy(result, m.transactions)
	return result
}

func (m *MockDB) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.transactions = make([]*model.Transaction, 0)
	m.insertError = nil
	m.pingError = nil
	m.closed = false
}
