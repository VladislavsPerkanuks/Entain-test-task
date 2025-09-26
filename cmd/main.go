package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

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
		return fmt.Errorf("Failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed successfully")

	return nil
}

func main() {
	serverConfig := config.DefaultConfig()

	dataSource := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		serverConfig.DB_USER, serverConfig.DB_PASSWORD, serverConfig.DB_HOST, serverConfig.DB_PORT, serverConfig.DB_NAME)

	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %s", err)
	}

	if err := migrateDB(db); err != nil {
		log.Fatalf("failed to migrate DB: %s", err)
	}

	transactionRepository := repository.NewRepository(db)
	transactionService := service.NewTransactionService(transactionRepository)

	router := httpServer.NewRouter(transactionService)

	log.Printf("Starting server on :%s...\n", serverConfig.SERVER_PORT)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverConfig.SERVER_PORT), router); err != nil {
		log.Printf("Server failed to start: %v\n", err)
	}
}
