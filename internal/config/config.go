package config

import (
	"fmt"
	"log"
	"os"
)

// Creating the necessary config information
type Config struct {
	DatabaseURL string
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

	log.Println("Connection Url is created");

	return &Config{
		DatabaseURL: databaseURL,
	}, nil
}
