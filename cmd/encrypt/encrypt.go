package main

import (
	"log"
	"os"
	"strings"
	"wa_chat_service/config"
	encrypt_service "wa_chat_service/internal/service/encrypt"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	zlogDevelopment := zap.NewDevelopmentConfig()
	zlogDevelopment.DisableCaller = true
	zlogDevelopment.DisableStacktrace = true
	zlogDevelopment.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zlogDevelopment.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	zlogDevelopment.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zlog, err := zlogDevelopment.Build()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	encryptService := encrypt_service.NewEncryptService(&config.Encrypt, zlog.Sugar())
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
