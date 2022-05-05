package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func MakeLogger(verbosity, encoding string) (*zap.Logger, error) {

	var level zapcore.Level

	switch verbosity {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	default:
		level = zapcore.ErrorLevel
	}

	cfg := zap.Config{
		Encoding:         encoding,
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.RFC3339TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	return cfg.Build()
}
