package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	logger, err := setupLogger(cfg.Logging.Level)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.New(ctx, cfg.Database.ConnString, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srv := server.New(db, logger)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv.Routes(),
	}

	go func() {
		// Set up metrics, TLS, graceful shutdown, etc
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	logger.Info("shutting down server")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown error", zap.Error(err))
	}
}

func setupLogger(levelStr string) (*zap.Logger, error) {
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

	defer func() {
		err := logger.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}()

	return logger, nil
}
