package server

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reimagined-pancake/server/auth"
	"time"

	"github.com/gin-gonic/gin"
)

// a basic Gin server with logging.

func ServeGin(db *sql.DB) {
	log.Println("Serving Gin.")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())

	v1 := r.Group("/v1")
	AddOpenRoutes(v1, db)

	protected := r.Group("/v1/api")
	protected.Use(auth.JWTAuth())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	config := fmt.Sprintf(":%s", port)
	log.Printf("Serving Gin on port :%s", port)

	r.Run(config)
}
