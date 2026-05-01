package users

import (
	"database/sql"
	"net/http"
	database "reimagined-pancake/db"
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
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
	}

	user.Password = hashedPassword

	err = database.SaveUser(db, ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "An error occured while saving database record"})
	}

	claims := jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key
	tokenString, err := token.SignedString(auth.JwtSecret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

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

	// Sign the token with our secret key
	tokenString, err := token.SignedString(auth.JwtSecret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":  userRecord,
		"token": tokenString,
	})
}
