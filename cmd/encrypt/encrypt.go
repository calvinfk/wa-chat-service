package main

import (
	"log"
	"os"
	"strings"
	"wa_chat_service/config"
	encrypt_service "wa_chat_service/internal/service/encrypt"

	"github.com/joho/godotenv"
)

func main() {
	args := os.Args
	if len(args) == 1 {
		log.Printf("[INFO][cmd/encrypt/encrypt.go][main] No command provided. Use 'encrypt' or 'decrypt'.")
		return
	}
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
	encryptService := encrypt_service.NewEncryptService(&config.Encrypt)
	option := strings.Split(args[1], "=")
	if len(option) != 2 {
		log.Printf("[INFO][cmd/encrypt/encrypt.go][main] Invalid command format. Use 'encrypt' or 'decrypt'.")
		return
	}
	switch option[0] {
	case "encrypt":
		encrypted, err := encryptService.Encrypt(option[1])
		if err != nil {
			log.Fatalf("Error encrypting: %v", err)
		}
		log.Printf("[INFO][cmd/encrypt/encrypt.go][main] Encrypted: %s", encrypted)
	case "decrypt":
		decrypted, err := encryptService.Decrypt(option[1])
		if err != nil {
			log.Fatalf("Error decrypting: %v", err)
		}
		log.Printf("[INFO][cmd/encrypt/encrypt.go][main] Decrypted: %s", decrypted)
	}

}
