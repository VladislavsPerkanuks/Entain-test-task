package http

import (
	"net/http"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/handler"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the HTTP router.
func NewRouter(transactionService service.TransactionService) chi.Router {
	handler := handler.NewHandler(transactionService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("OK"))
	})

	r.Group(func(r chi.Router) {
		r.Get("/user/{userID}/balance", handler.GetBalance)
		r.Post("/user/{userID}/transaction", handler.ProcessTransaction)
	})

	return r
}
