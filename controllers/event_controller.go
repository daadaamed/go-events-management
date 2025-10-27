package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/daadaamed/goeventmanagement/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Event struct {
	ID        int64                  `json:"id"`
	Source    string                 `json:"source"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Count     int64                  `json:"count"`
}

// GET /events
func GetEvents(c *gin.Context) {
	col := utils.DB.Collection("events")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	cur, err := col.Find(ctx, bson.D{}, options.Find().SetLimit(50).SetSort(bson.D{{Key: "timestamp", Value: -1}}))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer cur.Close(ctx)

	var out []Event
	if err := cur.All(ctx, &out); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// TODO : add POST /events
