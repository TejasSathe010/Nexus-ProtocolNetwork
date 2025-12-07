package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tejassathe/Nexus-ProtocolNetwork/internal/gateway"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/config"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/events"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

func main() {
	cfg := config.Load()

	logr := logger.New(cfg.LogLevel)

	logr.Info("starting omni-gateway",
		"listen_addr", cfg.ListenAddr,
		"env", cfg.Env,
	)

	eventService := events.NewLogService(logr)

	app := gateway.NewApp(cfg, logr, eventService)

	if err := run(app, logr, cfg); err != nil {
		logr.Error("gateway exited with error", "err", err)
		os.Exit(1)
	}
}

func run(app *gateway.App, logr logger.Logger, cfg config.Config) error {
	go func() {
		if err := app.Start(); err != nil {
			logr.Error("http server error", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logr.Info("shutting down omni-gateway")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return app.Shutdown(ctx)
}
