package appcontext

import (
	"go-leaderboard-server/internal/config"
	"go-leaderboard-server/internal/services"
)

type AppContext struct {
	AppConfig *config.Config
	services.Services
}
