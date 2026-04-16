package main

import (
	"log"
	"os"
	"wa_chat_service/config"
	"wa_chat_service/internal/app"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	config, err := config.New()
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	var zlog *zap.Logger
	if config.App.Environment == "production" {
		zlog, err = zap.NewProduction()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
	} else {
		zlogDevelopment := zap.NewDevelopmentConfig()
		zlogDevelopment.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		zlogDevelopment.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		zlogDevelopment.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		// Keep ANSI color in terminal output.
		stdoutEncoderConfig := zlogDevelopment.EncoderConfig
		stdoutEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		stdoutCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(stdoutEncoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)

		// Write plain text levels to file without ANSI color escape sequences.
		fileEncoderConfig := zlogDevelopment.EncoderConfig
		fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		fileCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(fileEncoderConfig),
			zapcore.AddSync(file),
			zapcore.DebugLevel,
		)

		zlog = zap.New(
			zapcore.NewTee(stdoutCore, fileCore),
			zap.AddCaller(),
		)
	}
	zslog := zlog.Sugar()
	app.Run(config, zslog)
}
