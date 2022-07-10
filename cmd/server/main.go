package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/bolom009/tcp-pow/internal/pkg/ctxlog"

	"github.com/bolom009/tcp-pow/internal/pkg/db"
	"go.uber.org/zap"

	"github.com/bolom009/tcp-pow/internal/pkg/config"
	"github.com/bolom009/tcp-pow/internal/server"
)

func main() {
	ctx := context.Background()

	logger := ctxlog.Logger(ctx, "main")
	logger.Info("StartServer")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("FailedToLoadConfig", zap.Error(err))
		return
	}

	redisCli, err := db.NewRedis(ctx, cfg.Redis.Host, cfg.Redis.Port)
	if err != nil {
		logger.Error("FailedToInitRedis", zap.Error(err))
		return
	}

	// seed for randomize quotes
	rand.Seed(time.Now().UnixNano())

	var (
		srv           = server.New(cfg, redisCli)
		serverAddress = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	)

	if err = srv.Start(ctx, serverAddress); err != nil {
		logger.Error("FailedToStartServer", zap.Error(err))
	}
}
