package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gerador-contratos/backend/internal/config"
	"gerador-contratos/backend/internal/db"
	apihttp "gerador-contratos/backend/internal/http"
	"gerador-contratos/backend/internal/http/handlers"
	"gerador-contratos/backend/internal/modules/auth"
	"gerador-contratos/backend/internal/modules/brokers"
	"gerador-contratos/backend/internal/modules/clauses"
	"gerador-contratos/backend/internal/modules/cnpj"
	"gerador-contratos/backend/internal/modules/contracts"
	"gerador-contratos/backend/internal/modules/rules"
)

func main() {
	cfg := config.Load()

	sqlDB, err := db.OpenSQLite(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()
	if err := db.RunMigrations(ctx, sqlDB, cfg.MigrationDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	authRepo := auth.NewSQLiteRepository(sqlDB)
	authService := auth.NewService(authRepo, auth.ServiceConfig{
		AccessTokenTTL:   cfg.AccessTokenTTL,
		RefreshTokenTTL:  cfg.RefreshTokenTTL,
		PasswordResetTTL: cfg.PasswordResetTokenTTL,
		JWTSecret:        cfg.JWTSecret,
		AppEnv:           cfg.AppEnv,
	})

	rulesService := rules.NewService()
	contractsRepo := contracts.NewSQLiteRepository(sqlDB)
	contractsService := contracts.NewService(contractsRepo, rulesService)

	brokersRepo := brokers.NewSQLiteRepository(sqlDB)
	brokersService := brokers.NewService(brokersRepo)

	clausesRepo := clauses.NewSQLiteRepository(sqlDB)
	clausesService := clauses.NewService(clausesRepo)

	cnpjRepo := cnpj.NewBrasilAPIRepository(nil, "")
	cnpjService := cnpj.NewService(cnpjRepo)

	router := apihttp.NewRouter(apihttp.RouterDependencies{
		AuthHandler:      handlers.NewAuthHandler(authService),
		ContractsHandler: handlers.NewContractsHandler(contractsService),
		BrokersHandler:   handlers.NewBrokersHandler(brokersService),
		ClausesHandler:   handlers.NewClausesHandler(clausesService),
		CNPJHandler:      handlers.NewCNPJHandler(cnpjService),
		AuthValidator:    authService,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("API running on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}
