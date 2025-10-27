package controllers

import (
	"context"
	"errors"
	"net/http"
	"time"

	service "github.com/daadaamed/goeventmanagement/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventHandler struct {
	svc service.EventService
}

func NewEventHandler(svc service.EventService) *EventHandler { return &EventHandler{svc: svc} }

func (h *EventHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/events", h.GetEvents)
	r.GET("events/:id", h.GetEventByID)
	r.POST("/events", h.PostEvent)
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
	event, err := h.svc.GetByID(ctx, c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) PostEvent(c *gin.Context) {
	var in service.EventIn
	if err := c.ShouldBindJSON(&in); err != nil || in.Source == "" || in.Type == "" || len(in.Payload) == 0 || string(in.Payload) == "null" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid body"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	ev, err := h.svc.Upsert(ctx, in)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ev)
}
