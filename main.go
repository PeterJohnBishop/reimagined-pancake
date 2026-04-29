package main

import (
	"log"
	"reimagined-pancake/server"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	server.ServeGin()
}
