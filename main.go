package main

import (
	"context"
	"errors"
	"log"
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

	go func() {
		// Set up metrics, TLS, graceful shutdown, etc
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit

}

func setupLogger(level string) (*log.Logger, error) {
	level, err := zapcore.ParseLevel(level)
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
