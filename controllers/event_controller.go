package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var events = []Event{
	{
		ID:        1,
		Source:    "app.web",
		Type:      "user_login",
		Timestamp: time.Date(2025, 1, 2, 15, 4, 5, 0, time.UTC),
		Payload:   map[string]interface{}{"user_id": 123, "ip": "1.2.3.4"},
		Count:     2,
	},
	{
		ID:        2,
		Source:    "worker.jobs",
		Type:      "job_completed",
		Timestamp: time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC),
		Payload:   map[string]interface{}{"job_id": "abc-42", "duration_ms": 840},
		Count:     1,
	},
}

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
	c.JSON(http.StatusOK, gin.H{
		"message": "list of events (placeholder)",
		"events":  events,
	})
}

// TODO : add POST /events
