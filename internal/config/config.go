package config

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

func DefaultConfig() *Config {
	return &Config{
		DB_HOST:     "localhost",
		DB_PORT:     "5432",
		DB_USER:     "postgres",
		DB_PASSWORD: "password",
		DB_NAME:     "database",
		SERVER_PORT: "3000",
	}
}
