package main

import (
	"io"
	"log"
	"os"
	"wa_chat_service/config"
	"wa_chat_service/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("APP_ENVIRONMENT") == "" || os.Getenv("APP_ENVIRONMENT") == "development" {
			log.Fatalf("Error loading .env file: %v, APP_ENVIRONMENT: %v", err, os.Getenv("APP_ENVIRONMENT"))
		}
	}

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// MultiWriter sends logs to both stdout and file
	multi := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multi)

	config, err := config.New()
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}
	app.Run(config)
}
