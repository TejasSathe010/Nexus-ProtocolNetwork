package control

import (
	"context"
	"net/http"
	"time"

	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/config"
	ctl "github.com/tejassathe/Nexus-ProtocolNetwork/pkg/control"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

type App struct {
	cfg        config.Config
	log        logger.Logger
	store      *ctl.Store
	httpServer *http.Server
}

func NewApp(cfg config.Config, log logger.Logger, store *ctl.Store) *App {
	router := NewRouter(log, store)

	srv := &http.Server{
		Addr:         cfg.ControlListenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		cfg:        cfg,
		log:        log,
		store:      store,
		httpServer: srv,
	}
}

func (a *App) Start() error {
	a.log.Info("control server starting", "addr", a.cfg.ControlListenAddr)
	return a.httpServer.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info("control server shutting down")
	return a.httpServer.Shutdown(ctx)
}
