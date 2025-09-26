package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/config"
	httpServer "github.com/VladislavsPerkanuks/Entain-test-task/internal/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func migrateDB(conf *config.Config) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.DB_USER, conf.DB_PASSWORD, conf.DB_HOST, conf.DB_PORT, conf.DB_NAME))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to the database: %v", err))
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to ping database: %v", err))
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to create postgres driver: %v", err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		panic(fmt.Sprintf("Failed to create migrate instance: %v", err))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}

	fmt.Println("Database migrations completed successfully")
}

func main() {
	serverConfig := config.DefaultConfig()

	migrateDB(serverConfig)

	router := httpServer.NewRouter()

	fmt.Printf("Starting server on :%s...\n", serverConfig.SERVER_PORT)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverConfig.SERVER_PORT), router); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
