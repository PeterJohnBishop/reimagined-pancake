package main

import (
	"context"
	"database/sql"
	"log"
	database "reimagined-pancake/database"
	"reimagined-pancake/server"

	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	db, err := database.InitDB(ctx, "data.db")
	server.ServeGin(db)
}
