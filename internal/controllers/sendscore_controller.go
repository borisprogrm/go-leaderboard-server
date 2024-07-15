package controllers

import (
	ac "go-leaderboard-server/internal/appcontext"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SendScoreParams struct {
	GameId string  `json:"gameId" binding:"required,max=50,alphanum" example:"game1" extensions:"x-order=0"`            // Id of game (alphanumeric values)
	UserId string  `json:"userId" binding:"required,max=50,alphanum" example:"user1" extensions:"x-order=1"`            // Id of user (alphanumeric values)
	Score  float64 `json:"score" binding:"required,min=0" example:"1500" extensions:"x-order=2"`                        // User score
	Name   string  `json:"name,omitempty" binding:"max=50" example:"John" extensions:"x-order=3"`                       // User name
	Params string  `json:"params,omitempty" binding:"max=255" example:"some additional payload" extensions:"x-order=4"` // Additional payload
}

// @Description Stores user data in a database
// @Tags user
// @Accept json
// @Produce json
// @Param data body SendScoreParams true "Body data"
// @Success 200 {object} ResultSuccess "Successful response"
// @Failure 400 {object} ResultError "Error response"
// @Failure 500 {object} ResultError "Error response"
// @Router /leaderboard/SendScore [put]
func SendScoreHandler(c *gin.Context) {
	var (
		ac     ac.AppContext = c.MustGet("appcontext").(ac.AppContext)
		params SendScoreParams
		err    error
		logger = log.GetLogger()
	)

	err = c.ShouldBindJSON(&params)
	if err != nil {
		logger.Error("Wrong params", log.LogParams{"error": err})
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = ac.LeaderboardService.PutUserScore(c, params.GameId, params.UserId, dbprovider.UserProperties{
		Score:  dbprovider.UScoreType(params.Score),
		Name:   params.Name,
		Params: params.Params,
	})
	if err != nil {
		logger.Error("Failed to put user score", log.LogParams{"error": err, "gameId": params.GameId, "userId": params.UserId})
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &ResultSuccess{Result: "success"})
}
