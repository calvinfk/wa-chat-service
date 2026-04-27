package main

import (
	"log"
	"os"
	"wa_chat_service/config"
	"wa_chat_service/internal/app"
	"wa_chat_service/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("APP_ENVIRONMENT") == "" || os.Getenv("APP_ENVIRONMENT") == "development" {
			log.Fatalf("Error loading .env file: %v, APP_ENVIRONMENT: %v", err, os.Getenv("APP_ENVIRONMENT"))
		}
	}

	config, err := config.New()
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	zlog, err := utils.NewZapLogger(config.App.Environment)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	zsLog := zlog.Sugar()
	app.Run(config, zsLog)
}
