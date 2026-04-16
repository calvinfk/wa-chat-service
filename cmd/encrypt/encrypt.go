package main

import (
	"log"
	"os"
	"strings"
	"wa_chat_service/config"
	encrypt_service "wa_chat_service/internal/service/encrypt"
	"wa_chat_service/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	args := os.Args
	if len(args) == 1 {
		log.Fatalf("[main] No command provided. Use 'encrypt' or 'decrypt'.")
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
	zlog, err := utils.NewZapLogger(config.App.Environment, nil)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	zslog := zlog.Sugar()
	encryptService := encrypt_service.NewEncryptService(&config.Encrypt, zslog)
	option := strings.Split(args[1], "=")
	if len(option) != 2 {
		log.Fatalf("[main] Invalid command format. Use 'encrypt' or 'decrypt'.")
	}
	switch option[0] {
	case "encrypt":
		encrypted, err := encryptService.Encrypt(option[1])
		if err != nil {
			zslog.Fatalf("[main] Error encrypting: %v", err)
		}
		zslog.Infof("[main] Encrypted: %s", encrypted)
	case "decrypt":
		decrypted, err := encryptService.Decrypt(option[1])
		if err != nil {
			zslog.Fatalf("[main] Error decrypting: %v", err)
		}
		zslog.Infof("[main] Decrypted: %s", decrypted)
	}

}
