package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
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
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	var fromParsedTime, toParsedTime *time.Time
	if s := c.Query("from"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			tt := t.UTC()
			fromParsedTime = &tt
		}
	}
	if s := c.Query("to"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			tt := t.UTC()
			toParsedTime = &tt
		}
	}

	query := service.ListQuery{
		Source: c.Query("source"),
		Type:   c.Query("type"),
		From:   fromParsedTime,
		To:     toParsedTime,
		Limit:  limit,
		Offset: offset,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	events, err := h.svc.List(ctx, query)
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
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) PostEvent(c *gin.Context) {
	var in service.EventIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format"})
		return
	}
	if in.Source == "" || in.Type == "" || len(in.Payload) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source, type and payload are required"})
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
