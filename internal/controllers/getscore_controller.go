package controllers

import (
	ac "go-leaderboard-server/internal/appcontext"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetScoreParams struct {
	GameId string `json:"gameId" binding:"required,max=50,alphanum" example:"game1" extensions:"x-order=0"` // Id of game (alphanumeric values)
	UserId string `json:"userId" binding:"required,max=50,alphanum" example:"user1" extensions:"x-order=1"` // Id of user (alphanumeric values)
}

type GetScoreResultSuccess[T any] struct {
	Result T `json:"result" binding:"required"` // (Empty object if no data)
}

// @Description Gets user data from a database
// @Tags user
// @Accept json
// @Produce json
// @Param data body GetScoreParams true "Body data"
// @Success 200 {object} GetScoreResultSuccess[dbprovider.UserProperties] "Successful response"
// @Failure 400 {object} ResultError "Error response"
// @Failure 500 {object} ResultError "Error response"
// @Router /leaderboard/GetScore [put]
func GetScoreHandler(c *gin.Context) {
	var (
		ac     ac.AppContext = c.MustGet("appcontext").(ac.AppContext)
		params GetScoreParams
		err    error
		logger = log.GetLogger()
	)

	err = c.ShouldBindJSON(&params)
	if err != nil {
		logger.Error("Wrong params", log.LogParams{"error": err})
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	userProp, err := ac.LeaderboardService.GetUserScore(c, params.GameId, params.UserId)
	if err != nil {
		logger.Error("Failed to get user score", log.LogParams{"error": err, "gameId": params.GameId, "userId": params.UserId})
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if userProp == nil {
		c.JSON(http.StatusOK, &GetScoreResultSuccess[struct{}]{})
		return
	}

	c.JSON(http.StatusOK, &GetScoreResultSuccess[dbprovider.UserProperties]{Result: *userProp})
}
