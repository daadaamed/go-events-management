package main

import (
	"github.com/daadaamed/goeventmanagement/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	utils.InitDB()

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Run()
}
