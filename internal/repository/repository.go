package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type Repository interface {
	// WithDBTransaction wraps the repository operations in a transaction
	WithDBTransaction(fn func(Repository) error) error

	// Balance Repository
	GetBalanceByID(userID int) (decimal.Decimal, error)
	UpdateUserBalance(userID int, delta decimal.Decimal) error

	// Transaction Repository
	GetTransactionByID(txID uuid.UUID) (*model.Transaction, error)
	InsertTransaction(tx *model.Transaction) error
}

type RepositoryImpl struct {
	db *sql.DB
	tx *sql.Tx
}

func NewRepository(db *sql.DB) Repository {
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) WithDBTransaction(fn func(Repository) error) error {
	if r.tx != nil { // avoid nested transactions
		return fn(r)
	}

	sqlTx, beginErr := r.db.Begin()
	if beginErr != nil {
		return fmt.Errorf("failed to begin transaction: %w", beginErr)
	}

	txRepo := &RepositoryImpl{db: r.db, tx: sqlTx}
	var fnErr error

	defer func() {
		if p := recover(); p != nil {
			_ = sqlTx.Rollback()
			panic(p)
		}

		if fnErr != nil {
			_ = sqlTx.Rollback()

			return
		}

		if commitErr := sqlTx.Commit(); commitErr != nil {
			fnErr = fmt.Errorf("failed to commit transaction: %w", commitErr)
		}
	}()

	fnErr = fn(txRepo)

	return fnErr
}

func (r *RepositoryImpl) GetBalanceByID(userID int) (decimal.Decimal, error) {
	var balance decimal.Decimal

	err := r.queryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balance for user %d: %w", userID, err)
	}

	return balance, nil
}

func (r *RepositoryImpl) UpdateUserBalance(userID int, delta decimal.Decimal) error {
	if _, err := r.exec(`
UPDATE users
SET balance = balance + $1
WHERE id = $2`, delta, userID); err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}

func (r *RepositoryImpl) GetTransactionByID(txID uuid.UUID) (*model.Transaction, error) {
	var tx model.Transaction

	err := r.queryRow(`
SELECT id, user_id, state, amount, source_type, created_at
FROM transactions
WHERE id = $1`, txID).Scan(&tx.ID, &tx.UserID, &tx.State, &tx.Amount, &tx.SourceType, &tx.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTransactionNotFound
		}

		return nil, fmt.Errorf("failed to find transaction by ID: %w", err)
	}

	return &tx, nil
}

func (r *RepositoryImpl) InsertTransaction(tx *model.Transaction) error {
	if _, err := r.exec(`
INSERT INTO transactions
(id, user_id, state, amount, source_type, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`,
		tx.ID, tx.UserID, tx.State, tx.Amount, tx.SourceType, tx.CreatedAt); err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	return nil
}

func (r *RepositoryImpl) exec(query string, args ...any) (sql.Result, error) {
	if r.tx != nil {
		return r.tx.Exec(query, args...)
	}

	return r.db.Exec(query, args...)
}

func (r *RepositoryImpl) queryRow(query string, args ...any) *sql.Row {
	if r.tx != nil {
		return r.tx.QueryRow(query, args...)
	}

	return r.db.QueryRow(query, args...)
}
