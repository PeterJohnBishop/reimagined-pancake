package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"reimagined-pancake/server/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StoreEventRequest struct {
	EventType string          `json:"event-type" binding:"required"`
	Data      json.RawMessage `json:"data" binding:"required"`
}

func StoreEventHandler(store *database.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StoreEventRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		event, err := store.StoreEvent(req.EventType, req.Data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store event"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Event stored",
			"event":   event,
		})
	}
}

func GetAllEventsHandler(store *database.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		events, err := store.GetAllEvents()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		}
		c.JSON(http.StatusOK, events)
	}
}

func GetEventByIDHandler(store *database.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")

		eventID, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		event, err := store.GetEventByID(eventID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, event)
	}
}

func GetEventsByTypeHandler(store *database.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		typeParam := c.Param("type")

		events, err := store.GetEventsByType(typeParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
			return
		}

		c.JSON(http.StatusOK, events)
	}
}

func DeleteEventHandler(store *database.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ID param muast not be empty."})
		}

		id, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		err = store.DeleteEvent(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Event successfully deleted"})

	}
}
