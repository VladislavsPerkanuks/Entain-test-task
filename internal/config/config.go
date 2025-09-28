package config

import "os"

type Config struct {
	// DB
	DB_HOST     string
	DB_PORT     string
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string

	// Server
	SERVER_PORT string
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		DB_HOST:     getEnvOrDefault("DB_HOST", "localhost"),
		DB_PORT:     getEnvOrDefault("DB_PORT", "5432"),
		DB_USER:     getEnvOrDefault("DB_USER", "postgres"),
		DB_PASSWORD: getEnvOrDefault("DB_PASSWORD", "password"),
		DB_NAME:     getEnvOrDefault("DB_NAME", "database"),
		SERVER_PORT: getEnvOrDefault("SERVER_PORT", "3000"),
	}
}
