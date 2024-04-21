package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load("local.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	botToken := os.Getenv("BOT_TOKEN")
}
