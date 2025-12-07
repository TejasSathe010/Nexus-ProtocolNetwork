package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/config"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/events"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/realtime"
)

type App struct {
	cfg        config.Config
	log        logger.Logger
	eventSvc   events.Service
	httpServer *http.Server
}

func NewApp(cfg config.Config, log logger.Logger, eventSvc events.Service) *App {
	wsHub := realtime.NewWSHub()
	sseBroker := realtime.NewSSEBroker()
	rtBroadcaster := realtime.NewBroadcaster(log, wsHub, sseBroker)

	router := NewRouter(log, eventSvc, wsHub, sseBroker, rtBroadcaster)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		cfg:        cfg,
		log:        log,
		eventSvc:   eventSvc,
		httpServer: srv,
	}
}

func (a *App) Start() error {
	a.log.Info("http server starting", "addr", a.cfg.ListenAddr)
	return a.httpServer.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info("http server shutting down")
	return a.httpServer.Shutdown(ctx)
}
