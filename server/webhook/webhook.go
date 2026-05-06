package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Payload struct {
	Data []byte
}

func WebhookHandler(payloadChan chan<- Payload) gin.HandlerFunc {
	const MaxBodyBytes = 25 << 20

	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			if err.Error() == "http: request body too large" {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"error": "Payload exceeds the 25MB limit",
				})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			return
		}

		if len(body) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Payload cannot be empty",
			})
			return
		}

		select {
		case payloadChan <- Payload{Data: body}:
			// 202 Accepted status for asynchronous processing
			c.JSON(http.StatusAccepted, gin.H{
				"status": "queued",
				"bytes":  len(body),
			})
		default:
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Server is currently too busy to accept new payloads",
			})
		}
	}
}

func ProcessPayloads(ch <-chan Payload, wg *sync.WaitGroup) {
	defer wg.Done()

	for p := range ch {
		var payload map[string]interface{}
		err := json.Unmarshal(p.Data, &payload)
		if err != nil {
			log.Printf("Error: failed to unmarshal the payload")
		}
		log.Printf("Processing payload asynchronously: %+v\n", payload)
		// switch on some event types or try to unmarshal
	}
	log.Println("Payload channel closed, background worker stopped.")
}
