package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/isaquebock/flags-api/internal/config"
	"github.com/isaquebock/flags-api/internal/redis"
	"github.com/isaquebock/flags-api/internal/server"
	"github.com/isaquebock/flags-api/internal/snapshot"
)

func main() {
	cfg := config.LoadFromEnv()

	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler)

	rdb, err := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	store := snapshot.NewRedisStore(rdb, cfg.SnapshotMaxRetries)

	srv := server.New(server.Deps{
		Logger:        logger,
		Store:         store,
		Config:        cfg,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Run(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
