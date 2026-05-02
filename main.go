package main

import (
	"database/sql"
	"log"
	"reimagined-pancake/server"

	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	server.ServeGin()
}
