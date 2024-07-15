package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Description Returns server status (success code)
// @Tags status
// @Produce json
// @Success 200 {object} ResultSuccess "Successful response"
// @Failure 500 {object} ResultError "Error response"
// @Router /Status [get]
func StatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, &ResultSuccess{Result: "success"})
}
