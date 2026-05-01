package server

import (
	"database/sql"
	"reimagined-pancake/server/users"

	"github.com/gin-gonic/gin"
)

func AddOpenRoutes(open *gin.RouterGroup, db *sql.DB) {
	open.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"serving": "gin",
		})
	})
	open.POST("/signup", func(ctx *gin.Context) {})
	open.POST("/login", func(ctx *gin.Context) {
		users.LoginHandler(ctx, db)
	})
}

func AddProtectedRoutes(protected *gin.RouterGroup, db *sql.DB) {
	//
}
