package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	BasicLogger *zap.Logger
	Logger      *zap.SugaredLogger
	logLevels   = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
	}
)

func init() {
	BasicLogger, _ = zap.NewProduction()
	Logger = BasicLogger.Sugar()
}

func InitLogger(serviceName, logLevel string) error {
	cfg := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(logLevels[logLevel]),
		OutputPaths: []string{"stdout"},
		InitialFields: map[string]interface{}{
			"service": serviceName,
		},
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}
	var err error
	BasicLogger, err = cfg.Build()
	if err != nil {
		return err
	}

	Logger = BasicLogger.Sugar()
	return nil
}
