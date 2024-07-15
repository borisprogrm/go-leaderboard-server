package controllers

import (
	ac "go-leaderboard-server/internal/appcontext"
	log "go-leaderboard-server/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeleteScoreParams struct {
	GameId string `json:"gameId" binding:"required,max=50,alphanum" example:"game1" extensions:"x-order=0"` // Id of game (alphanumeric values)
	UserId string `json:"userId" binding:"required,max=50,alphanum" example:"user1" extensions:"x-order=1"` // Id of user (alphanumeric values)
}

// @Description Removes user data from a database
// @Tags user
// @Accept json
// @Produce json
// @Param data body DeleteScoreParams true "Body data"
// @Success 200 {object} ResultSuccess "Successful response"
// @Failure 400 {object} ResultError "Error response"
// @Failure 500 {object} ResultError "Error response"
// @Router /leaderboard/DeleteScore [put]
func DeleteScoreHandler(c *gin.Context) {
	var (
		ac     ac.AppContext = c.MustGet("appcontext").(ac.AppContext)
		params DeleteScoreParams
		err    error
		logger = log.GetLogger()
	)

	err = c.ShouldBindJSON(&params)
	if err != nil {
		logger.Error("Wrong params", log.LogParams{"error": err})
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = ac.LeaderboardService.DeleteUserScore(c, params.GameId, params.UserId)
	if err != nil {
		logger.Error("Failed to delete user score", log.LogParams{"error": err, "gameId": params.GameId, "userId": params.UserId})
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &ResultSuccess{Result: "success"})
}
