package main

import (
	"context"
	"fmt"
	"log"

	"github.com/arbazshaikh150/TimeCourt/internal/config"
	"github.com/arbazshaikh150/TimeCourt/internal/database"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/arbazshaikh150/TimeCourt/internal/server"
)

func main() {
	// Making a postgress connection
	cfg, err := config.Load()
	ctx := context.Background()
	if err != nil {
		log.Fatal("Error Reading a database config from the environment variable", err)
	}
	// Making a connection to postgreSQL
	fmt.Println("Establishing the postgres Connection")
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)

	if err != nil {
		log.Fatal("Error in creating the database connection", err)
	}
	defer pool.Close()

	fmt.Println("Connection to postgress is successfull")
	gormDB, err := database.NewGorm(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Doing auto migration
	err = gormDB.AutoMigrate(
		&model.DecisionRecord{},
		&model.DecisionDetails{},
		&model.DecisionVersion{},
		&model.FactInformation{},
		&model.FactVersion{},
		&model.Idempotent{},
		&model.ResolutionVersion{},
		&model.ResolutionTable{},
		&model.RuleDetail{},
		&model.RuleTable{},
		&model.RuleVersion{},
	)
	if err != nil {
		log.Fatal("database migration failed:", err)
	}

	fmt.Println("Database migration successful")

	// Creating a server
	fmt.Println("Creating the server")
	httpServer := server.NewServer(cfg.ServerPort)

	fmt.Println("Starting TimeCourt server")
	fmt.Printf("Server is listening at port: %s\n", cfg.ServerPort)

	if err := httpServer.Start(); err != nil {
		log.Fatal("HTTP server stopped:", err)
	}

}
