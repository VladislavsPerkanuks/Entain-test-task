package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/model"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/service"
	"github.com/go-chi/chi/v5"
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
		"user_id": userID,
		"balance": balance,
	}

	json.NewEncoder(w).Encode(response)
}

func validateTransactionRequest(r *http.Request) (model.Transaction, error) {
	userID, err := validateUserID(r)
	if err != nil {
		return model.Transaction{}, err
	}

	transaction := model.Transaction{
		UserID: userID,
	}

	switch r.Header.Get("Source-Type") {
	case string(model.SourceTypeGame):
		transaction.SourceType = model.SourceTypeGame
	case string(model.SourceTypePayment):
		transaction.SourceType = model.SourceTypePayment
	case string(model.SourceTypeServer):
		transaction.SourceType = model.SourceTypeServer
	default:
		return model.Transaction{}, errors.New("invalid source type")
	}

	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		return model.Transaction{}, errors.New("invalid request")
	}

	if transaction.Amount.LessThanOrEqual(decimal.Zero) {
		return model.Transaction{}, errors.New("amount must be greater than zero")
	}

	if transaction.State != model.TransactionStateWin && transaction.State != model.TransactionStateLose {
		return model.Transaction{}, errors.New("invalid transaction state")
	}

	return transaction, nil
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
