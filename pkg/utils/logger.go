package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(environment string, file *os.File) (*zap.Logger, error) {
	var zlog *zap.Logger
	var err error
	if environment == "production" {
		zlog, err = zap.NewProduction()
		if err != nil {
			return nil, err
		}
	} else {
		var cores []zapcore.Core
		zlogDevelopment := zap.NewDevelopmentEncoderConfig()
		zlogDevelopment.EncodeLevel = zapcore.CapitalLevelEncoder
		zlogDevelopment.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		zlogDevelopment.EncodeCaller = zapcore.ShortCallerEncoder

		// Keep ANSI color in terminal output.
		stdoutEncoderConfig := zlogDevelopment
		stdoutEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		stdoutCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(stdoutEncoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
		cores = append(cores, stdoutCore)

		if file != nil {
			fileEncoderConfig := zlogDevelopment
			fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
			fileCore := zapcore.NewCore(
				zapcore.NewConsoleEncoder(fileEncoderConfig),
				zapcore.AddSync(file),
				zapcore.DebugLevel,
			)

			cores = append(cores, fileCore)
		}

		zlog = zap.New(
			zapcore.NewTee(cores...),
			zap.AddCaller(),
		)
	}
	return zlog, err
}
