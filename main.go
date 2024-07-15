package main

import (
	"go-leaderboard-server/internal/config"
	"go-leaderboard-server/internal/server"
	"os"

	log "go-leaderboard-server/internal/logger"
)

// @title Leaderboard API
// @version 1.0.0
// @contact.name borisprogrm
// @externalDocs.description OpenAPI
// @externalDocs.url https://swagger.io/resources/open-api/
func main() {
	config := config.GetAppConfig()

	logger := log.GetLogger()
	logger.Initialize(config.IsDebug)

	server := server.NewAppServer(nil)

	defer func() {
		if server.GetLastError() != nil {
			logger.Fatal("App exit with code 1")
		}
		logger.Info("App exit with code 0")
		os.Exit(0)
	}()

	defer func() {
		_ = server.Shutdown()
	}()

	err := server.Initialize(config)
	if err != nil {
		return
	}

	server.Start()
}
