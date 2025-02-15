package routes

import (
	"neobase-ai/internal/dtos"

	"github.com/gin-gonic/gin"
)

func SetupDefaultRoutes(router *gin.Engine) {
	// Health check route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, dtos.Response{
			Success: true,
			Data:    "Hello, World!",
		})
	})

}
