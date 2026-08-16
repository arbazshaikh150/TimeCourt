package config

import (
	"fmt"
	"os"
)

// Creating the necessary config information
type Config struct {
	DatabaseURL string
	ServerPort  string
}

func Load() (*Config, error) {
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	database := os.Getenv("DB_NAME")

	if username == "" ||
		password == "" ||
		host == "" ||
		port == "" ||
		database == "" {
		return nil, fmt.Errorf("database configuration is incomplete")
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		username,
		password,
		host,
		port,
		database,
	)

	port = os.Getenv("SERVER_PORT")

	return &Config{
		DatabaseURL: databaseURL,
		ServerPort:  port,
	}, nil
}
