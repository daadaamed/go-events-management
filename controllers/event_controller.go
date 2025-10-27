package controllers

import (
	"context"
	"net/http"
	"time"

	service "github.com/daadaamed/goeventmanagement/services"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	svc service.EventService
}

func NewEventHandler(svc service.EventService) *EventHandler { return &EventHandler{svc: svc} }

func (h *EventHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/events", h.GetEvents)
	r.GET("events/id")
}

// GET /events
func (h *EventHandler) GetEvents(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	events, err := h.svc.List(ctx, 50)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// GET BY ID /envet:id
func (h *EventHandler) GetEventByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	id := c.Param("id")
	event, err := h.svc.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "event not found" || err.Error() == "invalid event ID" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// TODO : add POST /events
