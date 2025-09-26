package repository

import (
	"database/sql"
	"fmt"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/model"
	"github.com/shopspring/decimal"
)

type Repository interface {
	GetBalance(userID int) (decimal.Decimal, error)
	ProcessTransaction(tx *model.Transaction) error
}

type RepositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) GetBalance(userID int) (decimal.Decimal, error) {
	var balance decimal.Decimal

	err := r.db.QueryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balance for user %d: %w", userID, err)
	}

	return balance, nil
}

func (r *RepositoryImpl) ProcessTransaction(tx *model.Transaction) error {
	sqlTx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			sqlTx.Rollback()
			panic(p)
		} else if err != nil {
			sqlTx.Rollback()
		} else {
			err = sqlTx.Commit()
		}
	}()

	_, err = sqlTx.Exec(`
INSERT INTO transactions
(id, user_id, state, amount, source_type, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`,
		tx.ID, tx.UserID, tx.State, tx.Amount, tx.SourceType, tx.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	// Update the user's balance
	var balanceUpdate decimal.Decimal

	switch tx.State {
	case model.TransactionStateWin:
		balanceUpdate = tx.Amount
	case model.TransactionStateLose:
		balanceUpdate = tx.Amount.Neg()
	default:
		return nil // Invalid state
	}

	_, err = sqlTx.Exec(`
UPDATE users
SET balance = balance + $1
WHERE id = $2`, balanceUpdate, tx.UserID)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	_, err = sqlTx.Exec(`
INSERT INTO processed_transactions
(transaction_id, user_id, processed_at)
VALUES ($1, $2, NOW())`, tx.ID, tx.UserID)
	if err != nil {
		return fmt.Errorf("failed to insert into processed_transactions: %w", err)
	}

	return nil
}
