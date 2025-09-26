package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID      uuid.UUID       `json:"id"`
	Balance decimal.Decimal `json:"balance"`
}

type TransactionState string

const (
	TransactionStateWin  TransactionState = "win"
	TransactionStateLose TransactionState = "lose"
)

type SourceType string

const (
	SourceTypeGame    SourceType = "game"
	SourceTypeServer  SourceType = "server"
	SourceTypePayment SourceType = "payment"
)

type Transaction struct {
	ID         uuid.UUID        `json:"id"`
	UserID     uuid.UUID        `json:"user_id"`
	State      TransactionState `json:"state"`
	Amount     decimal.Decimal  `json:"amount"`
	SourceType SourceType       `json:"source_type"`
	CreatedAt  time.Time        `json:"created_at"`
}

type ProcessedTransaction struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	UserID        uuid.UUID `json:"user_id"`
	ProcessedAt   time.Time `json:"processed_at"`
}
