package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/model"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionService interface {
	GetBalance(ctx context.Context, userID int) (decimal.Decimal, error)
	ProcessTransaction(ctx context.Context, tx *model.Transaction) error
}

type TransactionServiceImpl struct {
	repo repository.Repository
}

func NewTransactionService(repo repository.Repository) TransactionService {
	return &TransactionServiceImpl{repo: repo}
}

func (s *TransactionServiceImpl) GetBalance(ctx context.Context, userID int) (decimal.Decimal, error) {
	return s.repo.GetBalanceByID(ctx, userID)
}

func (s *TransactionServiceImpl) ProcessTransaction(ctx context.Context, tx *model.Transaction) error {
	_, err := s.repo.GetTransactionByID(ctx, tx.ID)
	if !errors.Is(err, repository.ErrTransactionNotFound) {
		return fmt.Errorf("transaction with ID %s already exists or failed to check existence: %w", tx.ID, err)
	}

	var balanceDelta decimal.Decimal

	switch tx.State {
	case model.TransactionStateWin:
		balanceDelta = tx.Amount
	case model.TransactionStateLose:
		balanceDelta = tx.Amount.Neg()
	default:
		return fmt.Errorf("unsupported transaction state: %s", tx.State)
	}

	if tx.ID == uuid.Nil {
		return errors.New("transaction ID cannot be nil")
	}

	return s.repo.WithDBTransaction(ctx, func(ctx context.Context, tr repository.Repository) error {
		if err = tr.InsertTransaction(ctx, tx); err != nil {
			return fmt.Errorf("failed to insert transaction: %w", err)
		}

		if err = tr.UpdateUserBalance(ctx, tx.UserID, balanceDelta); err != nil {
			return fmt.Errorf("failed to update user balance: %w", err)
		}

		return nil
	})
}
