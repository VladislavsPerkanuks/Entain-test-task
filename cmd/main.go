package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/config"
	httpServer "github.com/VladislavsPerkanuks/Entain-test-task/internal/http"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/repository"
	"github.com/VladislavsPerkanuks/Entain-test-task/internal/service"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func migrateDB(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")

	return nil
}

func main() {
	serverConfig := config.DefaultConfig()
	hostPort := net.JoinHostPort(serverConfig.DatabaseHost, serverConfig.DatabasePort)

	dataSource := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		serverConfig.DatabaseUser,
		serverConfig.DatabasePassword,
		hostPort,
		serverConfig.DatabaseName,
	)

	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Test the connection
	if err = db.PingContext(context.Background()); err != nil {
		log.Fatalf("failed to ping DB: %s", err)
	}

	if err = migrateDB(db); err != nil {
		log.Fatalf("failed to migrate DB: %s", err)
	}

	transactionRepository := repository.NewRepository(db)
	transactionService := service.NewTransactionService(transactionRepository)

	router := httpServer.NewRouter(transactionService)

	stop := make(chan os.Signal, 1)
	defer signal.Stop(stop)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", serverConfig.ServerPort),
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting server on :%s...\n", serverConfig.ServerPort)
		if err = server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}
	}()

	// Wait for a signal
	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
