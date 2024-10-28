package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samwestmoreland/transaction-store/internal/config"
	"github.com/samwestmoreland/transaction-store/internal/database/postgres"
	"github.com/samwestmoreland/transaction-store/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatal(err)
	}

	logger, err := setUpLogger(cfg.Logging.Level)
	defer func() {
		err := logger.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}()

	connString, err := getDBConnString()
	if err != nil {
		logger.Fatal("failed to get database connection string", zap.Error(err))
	}

	logger.Debug("database connection string", zap.String("conn_string", connString))

	dbConnCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.New(dbConnCtx, connString, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srv := server.New(db, logger)

	logger.Info("starting server", zap.Int("port", cfg.Server.Port))

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: srv.Routes(),
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown error", zap.Error(err))
	}
}

func setUpLogger(levelStr string) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(levelStr)
	if err != nil {
		return nil, err
	}

	logConfig := zap.NewProductionConfig()
	logConfig.Level = zap.NewAtomicLevelAt(level)

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func loadDBCredentialsFromEnv() (string, string, string, error) {
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return "", "", "", errors.New("POSTGRES_USER must be set")
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return "", "", "", errors.New("POSTGRES_PASSWORD must be set")
	}

	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		return "", "", "", errors.New("POSTGRES_DB must be set")
	}

	return user, password, dbName, nil
}

func getDBConnString() (string, error) {
	user, password, dbName, err := loadDBCredentialsFromEnv()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgresql://%s:%s@db:5432/%s", user, password, dbName), nil
}
