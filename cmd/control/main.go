package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tejassathe/Nexus-ProtocolNetwork/internal/control"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/config"
	ctl "github.com/tejassathe/Nexus-ProtocolNetwork/pkg/control"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/store"
)

func main() {
	cfg := config.Load()
	logr := logger.New(cfg.LogLevel)

	db, err := store.Open(cfg.DBDSN)
	if err != nil {
		logr.Error("failed to open db", "err", err)
		os.Exit(1)
	}
	if err := store.Migrate(db); err != nil {
		logr.Error("failed to migrate db", "err", err)
		os.Exit(1)
	}

	ctrlStore := ctl.NewStore(db)

	app := control.NewApp(cfg, logr, ctrlStore)

	if err := run(app, logr); err != nil {
		logr.Error("control service exited with error", "err", err)
		os.Exit(1)
	}
}

func run(app *control.App, logr logger.Logger) error {
	go func() {
		if err := app.Start(); err != nil {
			logr.Error("control http server error", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logr.Info("shutting down control service")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return app.Shutdown(ctx)
}
