package middleware

import (
	ac "go-leaderboard-server/internal/appcontext"
	"go-leaderboard-server/internal/controllers"

	"github.com/gin-gonic/gin"
)

func AppContextMiddleware(appContext *ac.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("appcontext", *appContext)
		c.Next()
	}
}

func ErrorHandlerMiddleware(isDebug bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			if isDebug {
				c.JSON(-1, &controllers.ResultError{Error: err.Error()})
			} else {
				var errMsg string // hide error details from client
				switch c.Writer.Status() {
				case 400:
					errMsg = "Wrong params"
				default:
					errMsg = "Internal server error"
				}
				c.JSON(-1, &controllers.ResultError{Error: errMsg})
			}
		}
	}
}
