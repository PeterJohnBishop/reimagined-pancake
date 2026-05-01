package users

import (
	"database/sql"
	"net/http"
	database "reimagined-pancake/database"
	"reimagined-pancake/global"
	"reimagined-pancake/server/auth"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func SignUpHandler(ctx *gin.Context, db *sql.DB) {

	var user global.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if user.Username == "" || user.Email == "" || user.Password == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "One or more required fields are empty."})
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}

	user.Password = hashedPassword

	err = database.SaveUser(db, ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "An error occured while saving database record"})
		return
	}

	claims := jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(auth.JwtSecret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.Password = ""

	ctx.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": tokenString,
	})

}

func LoginHandler(ctx *gin.Context, db *sql.DB) {
	var user global.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userRecord, err := database.GetUserByEmail(db, ctx, user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find a user with that email address"})
		return
	}

	passwordCheck := auth.CheckPasswordHash(user.Password, userRecord.Password)

	if user.Username != userRecord.Username || passwordCheck != true {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	claims := jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(auth.JwtSecret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	userRecord.Password = ""

	ctx.JSON(http.StatusOK, gin.H{
		"user":  userRecord,
		"token": tokenString,
	})
}

func GetUserHandler(ctx *gin.Context, db *sql.DB) {
	username := ctx.Param("username")

	userRecord, err := database.GetUserByUsername(db, ctx, username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	userRecord.Password = ""

	ctx.JSON(http.StatusOK, userRecord)
}

func UpdateUserHandler(ctx *gin.Context, db *sql.DB) {
	var updateData global.User
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	username := ctx.Param("username")
	existingUser, err := database.GetUserByUsername(db, ctx, username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if updateData.Email != "" {
		existingUser.Email = updateData.Email
	}
	if updateData.Password != "" {
		hashedPassword, _ := auth.HashPassword(updateData.Password)
		existingUser.Password = hashedPassword
	}

	err = database.UpdateUserByID(db, ctx, existingUser.ID, updateData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUserHandler(ctx *gin.Context, db *sql.DB) {
	username := ctx.Param("username")

	err := database.DeleteUserByID(db, ctx, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User account deleted"})
}
