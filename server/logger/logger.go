package logger

import (
	"go.uber.org/zap"

	"github.com/saiya/dsps/server/config"
)

// Logger is an interface of logger in this project
type Logger *zap.Logger

// DebugLogEnabler is an function to enable DEBUG level log on runtime
type DebugLogEnabler func()

// NewLogger initializes Logger
func NewLogger(config *config.ServerConfig) (Logger, DebugLogEnabler, error) {
	cfg := zap.NewProductionConfig()

	var dle DebugLogEnabler = func() {
		cfg.Level.SetLevel(zap.DebugLevel)
	}
	if config.Logging.Debug {
		dle()
	}

	fields := []zap.Field{}
	for key, value := range config.Logging.Attributes {
		fields = append(fields, zap.String(key, value))
	}

	logger, err := cfg.Build(zap.Fields(fields...))
	if err != nil {
		return nil, nil, err
	}
	return logger, dle, nil
}
