package routers

import (
	ac "go-leaderboard-server/internal/appcontext"
	"go-leaderboard-server/internal/controllers"
	log "go-leaderboard-server/internal/logger"
	"go-leaderboard-server/internal/middleware"
	"net/http"

	_ "go-leaderboard-server/docs"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

var logger = log.GetLogger()

func errorHandler(c *gin.Context, err any) {
	goErr := errors.Wrap(err, 3)
	logger.Error("Unexpected error", log.LogParams{"error": goErr, "stack": string(goErr.Stack())})
	c.AbortWithStatusJSON(http.StatusInternalServerError, &controllers.ResultError{Error: "Internal server error"})
}

func SetupRouter(appContext *ac.AppContext) *gin.Engine {
	if appContext.AppConfig.IsDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(gin.CustomRecovery(errorHandler))
	router.Use(middleware.AppContextMiddleware(appContext))
	router.Use(middleware.ErrorHandlerMiddleware(appContext.AppConfig.IsDebug))

	router.GET("/Status", controllers.StatusHandler)
	ldbrdGr := router.Group("/leaderboard")
	{
		ldbrdGr.POST("/SendScore", controllers.SendScoreHandler)
		ldbrdGr.POST("/DeleteScore", controllers.DeleteScoreHandler)
		ldbrdGr.POST("/GetScore", controllers.GetScoreHandler)
		ldbrdGr.POST("/GetTop", controllers.GetTopHandler)
	}

	if appContext.AppConfig.ApiUI {
		router.GET("/ui/*all", ginswagger.WrapHandler(swaggerfiles.Handler,
			ginswagger.DefaultModelsExpandDepth(-1)),
		)
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, &controllers.ResultError{Error: "Not found"})
	})

	return router
}
