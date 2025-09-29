package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTransactionService struct {
	mock.Mock

	processed []*model.Transaction
}

func (m *MockTransactionService) GetBalance(_ context.Context, userID int) (decimal.Decimal, error) {
	args := m.Called(userID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockTransactionService) ProcessTransaction(_ context.Context, tx *model.Transaction) error {
	args := m.Called(tx)
	err := args.Error(0)
	if err == nil {
		m.processed = append(m.processed, tx)
	}

	return err
}

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name    string
		param   string
		want    int
		wantErr bool
	}{
		{name: "valid integer", param: "42", want: 42, wantErr: false},
		{name: "negative integer", param: "-5", want: -5, wantErr: true},
		{name: "empty param", param: "", want: 0, wantErr: true},
		{name: "non-numeric", param: "abc", want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/dummy", nil)
			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("userID", tt.param)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			got, err := validateUserID(req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateTransactionRequest(t *testing.T) {
	transactionID := uuid.New().String()

	tests := []struct {
		name         string
		userParam    string
		sourceHeader string
		body         string
		wantErr      bool
	}{
		{
			name:         "valid win transaction",
			userParam:    "10",
			sourceHeader: string(model.SourceTypeGame),
			body:         `{"state":"win","amount":"50.00","transactionId":"` + transactionID + `"}`,
			wantErr:      false,
		},
		{
			name:         "valid lose transaction",
			userParam:    "7",
			sourceHeader: string(model.SourceTypePayment),
			body:         `{"state":"lose","amount":"25.50","transactionId":"` + uuid.New().String() + `"}`,
			wantErr:      false,
		},
		{
			name:         "invalid user ID",
			userParam:    "0",
			sourceHeader: string(model.SourceTypeGame),
			body:         `{"state":"win","amount":"50.00","transactionId":"` + uuid.New().String() + `"}`,
			wantErr:      true,
		},
		{
			name:         "invalid source header",
			userParam:    "5",
			sourceHeader: "unknown",
			body:         `{"state":"win","amount":"50.00","transactionId":"` + uuid.New().String() + `"}`,
			wantErr:      true,
		},
		{
			name:         "invalid json body",
			userParam:    "3",
			sourceHeader: string(model.SourceTypeServer),
			body:         `{"invalid`,
			wantErr:      true,
		},
		{
			name:         "invalid transaction id",
			userParam:    "3",
			sourceHeader: string(model.SourceTypeServer),
			body:         `{"state":"win","amount":"10.00","transactionId":"not-a-uuid"}`,
			wantErr:      true,
		},
		{
			name:         "zero amount",
			userParam:    "5",
			sourceHeader: string(model.SourceTypeGame),
			body:         `{"state":"win","amount":"0.00","transactionId":"` + uuid.New().String() + `"}`,
			wantErr:      true,
		},
		{
			name:         "invalid transaction state",
			userParam:    "5",
			sourceHeader: string(model.SourceTypeGame),
			body:         `{"state":"invalid","amount":"50.00","transactionId":"` + uuid.New().String() + `"}`,
			wantErr:      true,
		},
		{
			name:         "missing transaction id",
			userParam:    "5",
			sourceHeader: string(model.SourceTypeGame),
			body:         `{"state":"win","amount":"50.00"}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/dummy", bytes.NewBufferString(tt.body))
			if tt.sourceHeader != "" {
				req.Header.Set(SourceTypeHeader, tt.sourceHeader)
			}

			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("userID", tt.userParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			got, err := validateTransactionRequest(req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, got.ID)

			expectedID, convErr := strconv.Atoi(tt.userParam)
			require.NoError(t, convErr)
			assert.Equal(t, expectedID, got.UserID)
		})
	}
}

func TestHandlerGetBalance(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		setupMock  func(m *MockTransactionService)
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name:   "success",
			userID: "1",
			setupMock: func(m *MockTransactionService) {
				m.On("GetBalance", 1).Return(decimal.RequireFromString("123.45"), nil)
			},
			wantStatus: http.StatusOK,
			wantBody: map[string]any{
				"userId":  float64(1),
				"balance": "123.45",
			},
		},
		{
			name:       "validation error - invalid user ID",
			userID:     "0",
			setupMock:  func(_ *MockTransactionService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name:   "service error - user not found",
			userID: "1",
			setupMock: func(m *MockTransactionService) {
				m.On("GetBalance", 1).Return(decimal.Zero, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &MockTransactionService{}
			tt.setupMock(service)

			h := NewHandler(service)

			req := httptest.NewRequest(http.MethodGet, "/user/"+tt.userID+"/balance", nil)
			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("userID", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			resp := httptest.NewRecorder()
			h.GetBalance(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
			if tt.wantBody != nil {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				var body map[string]any
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
				assert.Equal(t, tt.wantBody["userId"], body["userId"])
				assert.Equal(t, tt.wantBody["balance"], body["balance"])
			}

			service.AssertExpectations(t)
		})
	}
}

func TestHandlerProcessTransaction(t *testing.T) {
	transactionID := uuid.New().String()

	tests := []struct {
		name          string
		requestBody   []byte
		userID        string
		sourceType    string
		setupMock     func(m *MockTransactionService)
		wantStatus    int
		wantProcessed bool
	}{
		{
			name:        "success",
			requestBody: []byte(`{"state":"win","amount":"10.00","transactionId":"` + transactionID + `"}`),
			userID:      "1",
			sourceType:  string(model.SourceTypeGame),
			setupMock: func(m *MockTransactionService) {
				m.On("ProcessTransaction", mock.AnythingOfType("*model.Transaction")).Return(nil)
			},
			wantStatus:    http.StatusOK,
			wantProcessed: true,
		},
		{
			name:          "bad request - invalid body",
			requestBody:   []byte(`{"state":"win"}`),
			userID:        "1",
			sourceType:    string(model.SourceTypeGame),
			setupMock:     func(_ *MockTransactionService) {},
			wantStatus:    http.StatusBadRequest,
			wantProcessed: false,
		},
		{
			name:        "service error - process failed",
			requestBody: []byte(`{"state":"lose","amount":"5.00","transactionId":"` + uuid.New().String() + `"}`),
			userID:      "1",
			sourceType:  string(model.SourceTypePayment),
			setupMock: func(m *MockTransactionService) {
				m.On("ProcessTransaction", mock.AnythingOfType("*model.Transaction")).
					Return(errors.New("insert failed"))
			},
			wantStatus:    http.StatusInternalServerError,
			wantProcessed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &MockTransactionService{}
			tt.setupMock(service)
			h := NewHandler(service)

			req := httptest.NewRequest(
				http.MethodPost,
				"/user/"+tt.userID+"/transaction",
				bytes.NewReader(tt.requestBody),
			)
			req.Header.Set(SourceTypeHeader, tt.sourceType)
			req.Header.Set("Content-Type", "application/json")

			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("userID", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			resp := httptest.NewRecorder()
			h.ProcessTransaction(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
			if tt.wantProcessed {
				require.Len(t, service.processed, 1)
				assert.Equal(t, 1, service.processed[0].UserID)
			} else {
				assert.Empty(t, service.processed)
			}

			service.AssertExpectations(t)
		})
	}
}
