package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"reimagined-pancake/server/webhook"
)

func setupWebhookRouter(ch chan<- webhook.Payload) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/webhook", webhook.WebhookHandler(ch))
	return r
}

func TestWebhookHandler(t *testing.T) {
	gofakeit.Seed(0)

	fmt.Printf("[ Webhook lifecycle test | Run ID: %s ]\n", gofakeit.UUID())

	t.Run("Happy Path - Successfully Queued", func(t *testing.T) {
		payloadChan := make(chan webhook.Payload, 1)
		router := setupWebhookRouter(payloadChan)

		testData := []byte(`{"event_type": "user_signup", "user_id": 123}`)
		req, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(testData))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Contains(t, w.Body.String(), "queued")

		select {
		case queuedPayload := <-payloadChan:
			assert.Equal(t, testData, queuedPayload.Data)
		default:
			t.Fatal("Expected payload in channel, but channel was empty")
		}
	})

	t.Run("Empty Payload Rejected", func(t *testing.T) {
		payloadChan := make(chan webhook.Payload, 1)
		router := setupWebhookRouter(payloadChan)

		req, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer([]byte{}))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Payload cannot be empty")
	})

	t.Run("3. Payload Exceeds 25MB Limit", func(t *testing.T) {
		payloadChan := make(chan webhook.Payload, 1)
		router := setupWebhookRouter(payloadChan)

		largePayloadSize := (25 << 20) + 1
		largePayload := make([]byte, largePayloadSize)

		req, _ := http.NewRequest("POST", "/webhook", bytes.NewReader(largePayload))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Contains(t, w.Body.String(), "Payload exceeds the 25MB limit")
	})

	t.Run("4. Channel Buffer Full (Too Many Requests)", func(t *testing.T) {
		payloadChan := make(chan webhook.Payload, 1)
		router := setupWebhookRouter(payloadChan)

		payloadChan <- webhook.Payload{Data: []byte("taking up the only slot")}

		testData := []byte(`{"event_type": "dropped_event"}`)
		req, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(testData))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "Server is currently too busy")
	})
}
