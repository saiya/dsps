package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
)

var rootLogger *loggerImpl
var globalLogLevel zap.AtomicLevel

func init() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Sampling = nil // Disable sampling

	globalLogLevel = cfg.Level
	zap, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(xerrors.Errorf("Failed to initialize zap logger: %w", err))
	}

	rootLogger = &loggerImpl{
		zap:    zap,
		filter: newDefaultFilter(),
	}
}

// EnableDebugLog enables debug log (process wide)
func EnableDebugLog() {
	globalLogLevel.SetLevel(zap.DebugLevel)
}

// InitLogger initializes Logger
func InitLogger(config *config.ServerConfig) error {
	if config.Logging.Debug {
		EnableDebugLog()
	}

	filter, err := NewFilter(config.Logging.Category)
	if err != nil {
		return err
	}

	fields := []zap.Field{}
	for key, value := range config.Logging.Attributes {
		fields = append(fields, zap.String(key, value))
	}

	rootLogger = rootLogger.WithFilter(filter).WithAttributes(fields)
	return nil
}
