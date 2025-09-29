package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/model"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Handler struct {
	ts service.TransactionService
}

func NewHandler(ts service.TransactionService) *Handler {
	return &Handler{ts: ts}
}

func validateUserID(r *http.Request) (int, error) {
	userID := chi.URLParam(r, "userID")

	userIDInt, err := strconv.Atoi(userID)
	if err != nil || userIDInt <= 0 {
		return 0, errors.New("invalid user ID")
	}

	return userIDInt, nil
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := validateUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balance, err := h.ts.GetBalance(userID)
	if err != nil {
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"userId": userID,
		"balance": balance.StringFixed(2),
	}

	json.NewEncoder(w).Encode(response)
}

type transactionRequestBody struct {
	State         string `json:"state"`
	Amount        string `json:"amount"`
	TransactionID string `json:"transactionId"`
}

const (
	SourceTypeHeader string = "Source-Type"
)

func validateTransactionRequest(r *http.Request) (model.Transaction, error) {
	userID, err := validateUserID(r)
	if err != nil {
		return model.Transaction{}, err
	}

	var reqBody transactionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return model.Transaction{}, errors.New("invalid request body")
	}

	amount, err := decimal.NewFromString(reqBody.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return model.Transaction{}, errors.New("amount must be a positive number")
	}

	state, err := model.ToTransactionState(reqBody.State)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("invalid transaction state: %w", err)
	}

	sourceType, err := model.ToSourceType(r.Header.Get(SourceTypeHeader))
	if err != nil {
		return model.Transaction{}, fmt.Errorf("invalid source type: %w", err)
	}

	transactionID, err := uuid.Parse(reqBody.TransactionID)
	if err != nil || transactionID == uuid.Nil {
		return model.Transaction{}, errors.New("invalid transactionId format")
	}

	return model.Transaction{
		ID:         transactionID,
		UserID:     userID,
		State:      state,
		Amount:     amount,
		SourceType: sourceType,
	}, nil
}

func (h *Handler) ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	validatedReq, err := validateTransactionRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.ts.ProcessTransaction(&validatedReq); err != nil {
		http.Error(w, "Failed to process transaction", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
