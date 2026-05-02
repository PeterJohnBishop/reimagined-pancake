package server

import (
	"net/http"
	"reimagined-pancake/server/database"

	"github.com/gin-gonic/gin"
)

func ServeGin() {

	database.ConnectPGDB()

	gin.Mode()
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"serving": "reimagined-pancakes with gin",
		})
	})

	r.Run()
}
