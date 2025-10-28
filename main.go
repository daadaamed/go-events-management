package main

import (
	"github.com/daadaamed/goeventmanagement/controllers"
	services "github.com/daadaamed/goeventmanagement/services"
	"github.com/daadaamed/goeventmanagement/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	client, db, err := utils.InitDB()
	if err != nil {
		panic(err)
	}
	defer func() { _ = client.Disconnect(nil) }()

	eventService := services.NewEventService(db)
	eventHandler := controllers.NewEventHandler(eventService)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	eventHandler.RegisterRoutes(router)
	// Add Swagger route
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
