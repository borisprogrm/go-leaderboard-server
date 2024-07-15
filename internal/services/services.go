package services

import (
	"context"
	"go-leaderboard-server/internal/config"
	log "go-leaderboard-server/internal/logger"
	"go-leaderboard-server/internal/utils"
	"time"
)

var logger = log.GetLogger()

type Services struct {
	LeaderboardService *LeaderboardService
}

func InitializeServices(ctx context.Context, config *config.Config, clock *utils.IClock, services *Services) error {
	var (
		err error
	)

	ctxInit, cancelInit := utils.GetContextByTimeout(ctx, time.Duration(config.TimeoutServicesInit)*time.Millisecond)
	defer cancelInit()

	services.LeaderboardService = NewLeaderboardService(config)
	err = services.LeaderboardService.Initialize(ctxInit, clock)

	return err
}

func ShutdownServices(ctx context.Context, config *config.Config, services *Services) error {
	var (
		err error
	)

	ctxShutdown, cancelShutdown := utils.GetContextByTimeout(ctx, time.Duration(config.TimeoutServicesShutdown)*time.Millisecond)
	defer cancelShutdown()

	if services.LeaderboardService != nil {
		err = services.LeaderboardService.Shutdown(ctxShutdown)
	}

	return err
}
