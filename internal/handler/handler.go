package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	ts service.TransactionService
}

func NewHandler(ts service.TransactionService) *Handler {
	return &Handler{ts: ts}
}

func validateBalanceRequest(r *http.Request) (int, error) {
	userID := chi.URLParam(r, "userID")

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return 0, errors.New("invalid user ID")
	}

	return userIDInt, nil
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := validateBalanceRequest(r)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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

func (h *Handler) ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement transaction processing logic
}
