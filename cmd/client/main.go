package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/bolom009/tcp-pow/internal/client"
	"github.com/bolom009/tcp-pow/internal/pkg/config"
	"github.com/bolom009/tcp-pow/internal/pkg/ctxlog"
)

func main() {
	ctx := context.Background()

	logger := ctxlog.Logger(ctx, "main")
	logger.Info("StartClient")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("FailedToLoadConfig", zap.Error(err))
		return
	}

	var (
		address = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		cli     = client.New(cfg)
	)

	if err = cli.Start(ctx, address); err != nil {
		logger.Error("FailedToStartClient", zap.Error(err))
	}
}
