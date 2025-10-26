package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /events
func GetEvents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "list of events (placeholder)",
	})
}
