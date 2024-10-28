package model

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID        uuid.UUID
	Amount    float64
	Timestamp time.Time
}
