package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID      int             `json:"id"`
	Balance decimal.Decimal `json:"balance"`
}

type TransactionState string

const (
	TransactionStateWin  TransactionState = "win"
	TransactionStateLose TransactionState = "lose"
)

func ToTransactionState(s string) (TransactionState, error) {
	switch s {
	case "win":
		return TransactionStateWin, nil
	case "lose":
		return TransactionStateLose, nil
	default:
		return "", fmt.Errorf("invalid transaction state: %s", s)
	}
}

type SourceType string

const (
	SourceTypeGame    SourceType = "game"
	SourceTypeServer  SourceType = "server"
	SourceTypePayment SourceType = "payment"
)

func ToSourceType(s string) (SourceType, error) {
	switch s {
	case "game":
		return SourceTypeGame, nil
	case "server":
		return SourceTypeServer, nil
	case "payment":
		return SourceTypePayment, nil
	default:
		return "", fmt.Errorf("invalid source type: %s", s)
	}
}

type Transaction struct {
	ID         uuid.UUID        `json:"transactionId"`
	UserID     int              `json:"userId"`
	State      TransactionState `json:"state"`
	Amount     decimal.Decimal  `json:"amount"`
	SourceType SourceType       `json:"sourceType"`
	CreatedAt  time.Time        `json:"createdAt"`
}

type ProcessedTransaction struct {
	TransactionID uuid.UUID `json:"transactionId"`
	UserID        int       `json:"userId"`
	ProcessedAt   time.Time `json:"processedAt"`
}
