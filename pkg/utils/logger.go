package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(environment string) (*zap.Logger, error) {
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
		zlogDevelopment.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		zlogDevelopment.EncodeCaller = zapcore.ShortCallerEncoder
		zlogDevelopment.EncodeLevel = zapcore.CapitalColorLevelEncoder
		stdoutCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(zlogDevelopment),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
		cores = append(cores, stdoutCore)

		zlog = zap.New(
			zapcore.NewTee(cores...),
			zap.AddCaller(),
		)
	}
	return zlog, err
}
