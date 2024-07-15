package server

import (
	"context"
	"errors"
	ac "go-leaderboard-server/internal/appcontext"
	"go-leaderboard-server/internal/config"
	log "go-leaderboard-server/internal/logger"
	"go-leaderboard-server/internal/routers"
	"go-leaderboard-server/internal/services"
	"go-leaderboard-server/internal/utils"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var logger = log.GetLogger()

type AppServer struct {
	clock      utils.IClock
	appContext *ac.AppContext
	router     *gin.Engine
	server     *http.Server
	err        error
}

func NewAppServer(clock utils.IClock) *AppServer {
	if clock == nil {
		clock = &utils.Clock{}
	}
	return &AppServer{
		clock: clock,
	}
}

func (s *AppServer) GetLastError() error {
	return s.err
}

func (s *AppServer) Initialize(config *config.Config) error {
	s.appContext = &ac.AppContext{
		AppConfig: config,
	}

	logger.Info("Server initialization...")

	s.err = services.InitializeServices(context.Background(), s.appContext.AppConfig, &s.clock, &s.appContext.Services)
	if s.err != nil {
		logger.Error("Failed to initialize server", log.LogParams{"error": s.err})
		return s.err
	}

	s.router = routers.SetupRouter(s.appContext)

	return nil
}

func (s *AppServer) Start() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	host := s.appContext.AppConfig.Host
	port := s.appContext.AppConfig.Port
	s.server = &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: s.router,
	}
	go func() {
		logger.Info("Server started", log.LogParams{"host": host, "port": port})
		s.err = s.server.ListenAndServe()
		if s.err != nil && s.err != http.ErrServerClosed {
			logger.Error("Failed to start server", log.LogParams{"error": s.err})
		} else {
			s.err = nil
		}
		stop()
	}()

	<-ctx.Done()
	stop()
}

func (s *AppServer) Shutdown() error {
	logger.Info("Server shutting down...")

	if s.server != nil {
		ctx, cancel := utils.GetContextByTimeout(
			context.Background(),
			time.Duration(s.appContext.AppConfig.TimeoutServerClose)*time.Millisecond,
		)
		defer cancel()
		err := s.server.Shutdown(ctx)
		if err != nil {
			logger.Error("Failed to shutdown server", log.LogParams{"error": s.err})
			s.err = errors.Join(s.err, err)
		}
	}

	if s.appContext != nil {
		err := services.ShutdownServices(context.Background(), s.appContext.AppConfig, &s.appContext.Services)
		if err != nil {
			logger.Error("Failed to shutdown services", log.LogParams{"error": err})
			s.err = errors.Join(s.err, err)
		}
	}

	return s.err
}
