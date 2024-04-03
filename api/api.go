package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Start() {
	router := gin.Default()

	router.GET("/", func(_context *gin.Context) {
		_context.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	router.Run("localhost:8080")
}
