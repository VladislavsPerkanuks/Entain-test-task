package http

import (
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/handler"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the HTTP router
func NewRouter(transactionService service.TransactionService) chi.Router {
	handler := handler.NewHandler(transactionService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/user/{userID}/balance", handler.GetBalance)
	r.Post("/user/{userID}/transaction", handler.ProcessTransaction)

	return r
}
