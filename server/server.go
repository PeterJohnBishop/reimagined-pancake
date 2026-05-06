package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"reimagined-pancake/server/database"
	"reimagined-pancake/server/utils"
	"reimagined-pancake/server/webhook"

	"github.com/gin-gonic/gin"
)

var store *database.DBStore

func ServeGin() {
	var err error
	store, err = database.ConnectPGDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database server: %v\n", err)
	}

	log.Println("Checking database tables...")
	if err := store.CreateTables(); err != nil {
		log.Fatalf("Failed to initialize database tables: %v\n", err)
	}
	log.Println("Database tables verified!")

	payloadChan := make(chan webhook.Payload, 100)
	var wg sync.WaitGroup

	wg.Add(1)
	go processPayloads(payloadChan, &wg)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"serving": "reimagined-pancakes with gin",
		})
	})

	r.POST("/webhook", webhook.WebhookHandler(payloadChan))

	protected := r.Group("/api")
	protected.Use(utils.RequireAuth())

	AddOpenRoutes(r, store)
	AddProtectedRoutes(protected, store)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("HTTP server stopped. Waiting for background tasks to complete...")

	close(payloadChan)

	wg.Wait()

	if store != nil && store.DB != nil {
		store.DB.Close()
	}

	log.Println("Server exiting")
}

func processPayloads(ch <-chan webhook.Payload, wg *sync.WaitGroup) {
	defer wg.Done()

	for p := range ch {
		var payload map[string]interface{}
		err := json.Unmarshal(p.Data, &payload)
		if err != nil {
			log.Printf("Error: failed to unmarshal the payload")
		}
		log.Printf("Processing payload asynchronously: %+v\n", payload)

	}
	log.Println("Payload channel closed, background worker stopped.")
}
