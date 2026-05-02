package server

import (
	"log"
	"net/http"
	"reimagined-pancake/server/database"
	"reimagined-pancake/server/utils"

	"github.com/gin-gonic/gin"
)

var store *database.DBStore

func ServeGin() {
	var err error
	store, err = database.ConnectPGDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database server: %v\n", err)
	}
	defer store.DB.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"serving": "reimagined-pancakes with gin",
		})
	})

	protected := r.Group("/api")
	protected.Use(utils.RequireAuth())

	AddOpenRoutes(r, store)
	AddProtectedRoutes(protected, store)

	r.Run()
}
