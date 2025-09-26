package service

import (
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/db"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/model"
	"github.com/shopspring/decimal"
)

type TransactionService interface {
	GetBalance(userID int) (decimal.Decimal, error)
	ProcessTransaction(tx *model.Transaction) error
}

type TransactionServiceImpl struct {
	repo db.Repository
}

func NewTransactionService(repo db.Repository) TransactionService {
	return &TransactionServiceImpl{repo: repo}
}

func (s *TransactionServiceImpl) GetBalance(userID int) (decimal.Decimal, error) {
	return s.repo.GetBalance(userID)
}

func (s *TransactionServiceImpl) ProcessTransaction(tx *model.Transaction) error {
	return s.repo.ProcessTransaction(tx)
}
