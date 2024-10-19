package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {

	// Callback
	r.POST("/callback", callbackHandler)
	r.POST("/echobot", echoBot)
}
