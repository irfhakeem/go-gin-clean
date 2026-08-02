package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var instance *zap.Logger

func Init(environment string) {
	var (
		l   *zap.Logger
		err error
	)

	if environment == "production" {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		l, err = cfg.Build()
	} else {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		l, err = cfg.Build()
	}

	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	instance = l
}

func Sync() {
	if instance != nil {
		_ = instance.Sync()
	}
}

func Get() *zap.Logger {
	return instance
}

func Info(msg string, fields ...zap.Field) {
	instance.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	instance.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	instance.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	instance.Fatal(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	instance.Debug(msg, fields...)
}
