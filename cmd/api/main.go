package main

import (
	"context"
	"fmt"
	"log"

	"github.com/arbazshaikh150/TimeCourt/internal/config"
	"github.com/arbazshaikh150/TimeCourt/internal/database"
)

func main() {
	// Making a postgress connection
	config, err := config.Load()
	context := context.Background()
	if err != nil {
		log.Fatal("Error Reading a database config from the environment variable", err)
	}
	// Making a connection to postgreSQL
	fmt.Println("Establishing the postgres Connection")
	db, err := database.NewPool(context, config.DatabaseURL)

	if err != nil {
		log.Fatal("Error in creating the database connection", err)
	}
	defer db.Close()

	fmt.Println("Connection to postgress is successfull")

}
