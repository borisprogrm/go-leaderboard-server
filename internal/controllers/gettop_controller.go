package controllers

import (
	ac "go-leaderboard-server/internal/appcontext"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetTopParams struct {
	GameId string `json:"gameId" binding:"required,max=50,alphanum" example:"game1" extensions:"x-order=0"` // Id of game (alphanumeric values)
	NTop   uint32 `json:"nTop" binding:"required,min=1,max=100" example:"100" extensions:"x-order=1"`       // Number of users in top
}

type GetTopResultSuccess struct {
	Result dbprovider.TopData `json:"result" binding:"required"`
}

// @Description Returns data of users with maximum registered scores sorted in descending order of score, maximum nTop number of elements for a specific gameId
// @Tags top
// @Accept json
// @Produce json
// @Param data body GetTopParams true "Body data"
// @Success 200 {object} GetTopResultSuccess "Successful response"
// @Failure 400 {object} ResultError "Error response"
// @Failure 500 {object} ResultError "Error response"
// @Router /leaderboard/GetTop [put]
func GetTopHandler(c *gin.Context) {
	var (
		ac     ac.AppContext = c.MustGet("appcontext").(ac.AppContext)
		params GetTopParams
		err    error
		logger = log.GetLogger()
	)

	err = c.ShouldBindJSON(&params)
	if err != nil {
		logger.Error("Wrong params", log.LogParams{"error": err})
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	top, err := ac.LeaderboardService.GetTop(c, params.GameId, params.NTop)
	if err != nil {
		logger.Error("Failed to get top", log.LogParams{"error": err, "gameId": params.GameId, "nTop": params.NTop})
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &GetTopResultSuccess{Result: top})
}
