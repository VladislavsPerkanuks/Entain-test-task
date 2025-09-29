package service

import (
	"context"
	"errors"
	"testing"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/model"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock

	balanceUpdates []struct {
		userID int
		delta  decimal.Decimal
	}
}

func (m *MockRepository) WithDBTransaction(
	ctx context.Context,
	fn func(context.Context, repository.Repository) error,
) error {
	args := m.Called(fn)
	// If the mock is configured to return an error for the transaction wrapper, return it.
	if err := args.Error(0); err != nil {
		return err
	}

	// Otherwise execute the provided function to simulate the transactional work.
	if fn == nil {
		return nil
	}

	return fn(ctx, m)
}

func (m *MockRepository) GetBalanceByID(_ context.Context, userID int) (decimal.Decimal, error) {
	args := m.Called(userID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockRepository) UpdateUserBalance(_ context.Context, userID int, delta decimal.Decimal) error {
	m.balanceUpdates = append(m.balanceUpdates, struct {
		userID int
		delta  decimal.Decimal
	}{userID: userID, delta: delta})
	args := m.Called(userID, delta)
	return args.Error(0)
}

func (m *MockRepository) GetTransactionByID(_ context.Context, id uuid.UUID) (*model.Transaction, error) {
	args := m.Called(id)
	tx, _ := args.Get(0).(*model.Transaction)
	return tx, args.Error(1)
}

func (m *MockRepository) InsertTransaction(_ context.Context, tx *model.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func TestProcessTransaction(t *testing.T) {
	winID := uuid.New()
	loseID := uuid.New()
	duplicateID := uuid.New()

	tests := []struct {
		name      string
		tx        *model.Transaction
		setupMock func(m *MockRepository)
		wantErr   bool
		wantDelta decimal.Decimal
	}{
		{
			name: "win",
			tx: &model.Transaction{
				ID:     winID,
				UserID: 1,
				State:  model.TransactionStateWin,
				Amount: decimal.NewFromInt(100),
			},
			setupMock: func(m *MockRepository) {
				m.On("WithDBTransaction", mock.Anything).Return(nil)
				m.On("GetTransactionByID", winID).Return(nil, repository.ErrTransactionNotFound)
				m.On("InsertTransaction", mock.Anything).Return(nil)
				m.On("UpdateUserBalance", 1, decimal.NewFromInt(100)).Return(nil)
			},
			wantErr:   false,
			wantDelta: decimal.NewFromInt(100),
		},
		{
			name: "lose",
			tx: &model.Transaction{
				ID:     loseID,
				UserID: 2,
				State:  model.TransactionStateLose,
				Amount: decimal.NewFromInt(50),
			},
			setupMock: func(m *MockRepository) {
				m.On("WithDBTransaction", mock.Anything).Return(nil)
				m.On("GetTransactionByID", loseID).Return(nil, repository.ErrTransactionNotFound)
				m.On("InsertTransaction", mock.Anything).Return(nil)
				m.On("UpdateUserBalance", 2, decimal.NewFromInt(50).Neg()).Return(nil)
			},
			wantErr:   false,
			wantDelta: decimal.NewFromInt(50).Neg(),
		},
		{
			name: "duplicate",
			tx: &model.Transaction{
				ID:     duplicateID,
				UserID: 1,
				State:  model.TransactionStateWin,
				Amount: decimal.NewFromInt(10),
			},
			setupMock: func(m *MockRepository) {
				m.On("GetTransactionByID", duplicateID).Return(&model.Transaction{ID: duplicateID}, nil)
			},
			wantErr: true,
		},
		{
			name: "insert error",
			tx: &model.Transaction{
				ID:     uuid.New(),
				UserID: 1,
				State:  model.TransactionStateWin,
				Amount: decimal.NewFromInt(10),
			},
			setupMock: func(m *MockRepository) {
				m.On("WithDBTransaction", mock.Anything).Return(nil)
				m.On("GetTransactionByID", mock.Anything).Return(nil, repository.ErrTransactionNotFound)
				m.On("InsertTransaction", mock.Anything).Return(errors.New("insert failed"))
			},
			wantErr: true,
		},
		{
			name: "update error",
			tx: &model.Transaction{
				ID:     uuid.New(),
				UserID: 1,
				State:  model.TransactionStateLose,
				Amount: decimal.NewFromInt(5),
			},
			setupMock: func(m *MockRepository) {
				m.On("WithDBTransaction", mock.Anything).Return(nil)
				m.On("GetTransactionByID", mock.Anything).Return(nil, repository.ErrTransactionNotFound)
				m.On("InsertTransaction", mock.Anything).Return(nil)
				m.On("UpdateUserBalance", 1, decimal.NewFromInt(5).Neg()).Return(errors.New("update failed"))
			},
			wantErr: true,
		},
		{
			name: "transaction wrapper error",
			tx: &model.Transaction{
				ID:     uuid.New(),
				UserID: 1,
				State:  model.TransactionStateWin,
				Amount: decimal.NewFromInt(5),
			},
			setupMock: func(m *MockRepository) {
				m.On("GetTransactionByID", mock.Anything).Return(nil, repository.ErrTransactionNotFound)
				m.On("WithDBTransaction", mock.Anything).Return(errors.New("tx fail"))
			},
			wantErr: true,
		},
		{
			name: "nil transaction id",
			tx:   &model.Transaction{UserID: 1, State: model.TransactionStateWin, Amount: decimal.NewFromInt(5)},
			setupMock: func(m *MockRepository) {
				m.On("GetTransactionByID", mock.Anything).Return(nil, repository.ErrTransactionNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			repo := &MockRepository{}
			tt.setupMock(repo)
			svc := NewTransactionService(repo)
			err := svc.ProcessTransaction(ctx, tt.tx)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, repo.balanceUpdates, 1)
				assert.True(t, repo.balanceUpdates[0].delta.Equal(tt.wantDelta))
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestGetBalance(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		setupMock func(m *MockRepository)
		wantErr   bool
		wantValue decimal.Decimal
	}{
		{
			name:   "success",
			userID: 1,
			setupMock: func(m *MockRepository) {
				m.On("GetBalanceByID", 1).Return(decimal.RequireFromString("42.50"), nil)
			},
			wantErr:   false,
			wantValue: decimal.RequireFromString("42.50"),
		},
		{
			name:   "repo error",
			userID: 2,
			setupMock: func(m *MockRepository) {
				m.On("GetBalanceByID", 2).Return(decimal.Zero, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			repo := &MockRepository{}
			tt.setupMock(repo)
			svc := NewTransactionService(repo)
			balance, err := svc.GetBalance(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, tt.wantValue.Equal(balance))
			}

			repo.AssertExpectations(t)
		})
	}
}
