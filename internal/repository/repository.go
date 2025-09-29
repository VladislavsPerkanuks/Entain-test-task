package repository

import (
	"context"
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
	WithDBTransaction(ctx context.Context, fn func(context.Context, Repository) error) error

	// Balance Repository
	GetBalanceByID(ctx context.Context, userID int) (decimal.Decimal, error)
	UpdateUserBalance(ctx context.Context, userID int, delta decimal.Decimal) error

	// Transaction Repository
	GetTransactionByID(ctx context.Context, txID uuid.UUID) (*model.Transaction, error)
	InsertTransaction(ctx context.Context, tx *model.Transaction) error
}

type Postgresql struct {
	db *sql.DB
	tx *sql.Tx
}

func NewRepository(db *sql.DB) Repository {
	return &Postgresql{db: db}
}

func (r *Postgresql) WithDBTransaction(ctx context.Context, fn func(context.Context, Repository) error) error {
	if r.tx != nil { // avoid nested transactions
		return fn(ctx, r)
	}

	sqlTx, beginErr := r.db.BeginTx(ctx, nil)
	if beginErr != nil {
		return fmt.Errorf("failed to begin transaction: %w", beginErr)
	}

	txRepo := &Postgresql{db: r.db, tx: sqlTx}
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

	fnErr = fn(ctx, txRepo)

	return fnErr
}

func (r *Postgresql) GetBalanceByID(ctx context.Context, userID int) (decimal.Decimal, error) {
	var balance decimal.Decimal

	err := r.queryRowContext(ctx, "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balance for user %d: %w", userID, err)
	}

	return balance, nil
}

func (r *Postgresql) UpdateUserBalance(ctx context.Context, userID int, delta decimal.Decimal) error {
	if _, err := r.exec(ctx, `
UPDATE users
SET balance = balance + $1
WHERE id = $2`, delta, userID); err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}

func (r *Postgresql) GetTransactionByID(ctx context.Context, txID uuid.UUID) (*model.Transaction, error) {
	var tx model.Transaction

	err := r.queryRowContext(ctx,
		`
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

func (r *Postgresql) InsertTransaction(ctx context.Context, tx *model.Transaction) error {
	if _, err := r.exec(ctx, `
INSERT INTO transactions
(id, user_id, state, amount, source_type)
VALUES ($1, $2, $3, $4, $5)`,
		tx.ID, tx.UserID, tx.State, tx.Amount, tx.SourceType); err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	return nil
}

func (r *Postgresql) exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if r.tx != nil {
		return r.tx.ExecContext(ctx, query, args...)
	}

	return r.db.ExecContext(ctx, query, args...)
}

func (r *Postgresql) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if r.tx != nil {
		return r.tx.QueryRowContext(ctx, query, args...)
	}

	return r.db.QueryRowContext(ctx, query, args...)
}
