package server

import (
	"reimagined-pancake/server/database"
	"reimagined-pancake/server/handlers"

	"github.com/gin-gonic/gin"
)

func AddOpenRoutes(open *gin.Engine, db *database.DBStore) {
	open.POST("/register", handlers.RegisterHandler(db))
	open.POST("/login", handlers.LoginHandler(db))
}

func AddProtectedRoutes(protected *gin.RouterGroup, db *database.DBStore) {
	protected.GET("/user/:id", handlers.GetUserHandler(db))
	protected.GET("/users", handlers.GetAllUsersHandler(db))
	protected.PUT("/user", handlers.UpdateProfileHandler(db))
	protected.PUT("/user/password", handlers.UpdatePasswordHandler(db))
	protected.DELETE("/user", handlers.DeleteUserHandler(db))
}
