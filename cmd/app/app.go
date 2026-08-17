package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arbazshaikh150/TimeCourt/internal/config"
	"github.com/arbazshaikh150/TimeCourt/internal/database"
	"github.com/arbazshaikh150/TimeCourt/internal/handler"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/arbazshaikh150/TimeCourt/internal/repository"
	"github.com/arbazshaikh150/TimeCourt/internal/router"
	"github.com/arbazshaikh150/TimeCourt/internal/server"
	"github.com/arbazshaikh150/TimeCourt/internal/service"
)

// App owns application startup and dependency wiring.
type App struct{}

func NewApp() *App {
	return &App{}
}

// Run initializes dependencies and starts the HTTP server.
func (a *App) Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	ctx := context.Background()
	fmt.Println("Establishing the postgres connection")
	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer db.Close()

	gormDB, err := database.NewGorm(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create gorm database connection: %w", err)
	}

	if err := gormDB.AutoMigrate(
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
	); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	// Wiring the dependencies
	factRepository := repository.NewFactInformationRepository(db.Pool)
	idempotentRepository := repository.NewIdempotentRepository(db.Pool)
	idempotentService := service.NewIdempotentService(idempotentRepository)
	factService := service.NewFactService(factRepository, *idempotentService)
	factHandler := handler.NewFactHandler(*factService)

	mux := http.NewServeMux()
	router.RegisterRoutes(mux, factHandler)

	httpServer := server.NewServer(cfg.ServerPort, mux)
	fmt.Printf("Server is listening at port %s\n", cfg.ServerPort)

	return httpServer.Start()
}
