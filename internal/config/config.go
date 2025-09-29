package config

import "os"

type Config struct {
	// DB
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string

	// Server
	ServerPort string
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		DatabaseHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DatabasePort:     getEnvOrDefault("DB_PORT", "5432"),
		DatabaseUser:     getEnvOrDefault("DB_USER", "postgres"),
		DatabasePassword: getEnvOrDefault("DB_PASSWORD", "password"),
		DatabaseName:     getEnvOrDefault("DB_NAME", "database"),
		ServerPort:       getEnvOrDefault("SERVER_PORT", "3000"),
	}
}
