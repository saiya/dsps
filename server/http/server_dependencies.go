package http

import (
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/util"
	"github.com/saiya/dsps/server/logger"
)

// ServerDependencies struct holds all resource references to build web server
type ServerDependencies struct {
	Logger          logger.Logger
	DebugLogEnabler logger.DebugLogEnabler

	ServerClose util.ServerClose

	Storage domain.Storage

	LongPollingMaxTimeout domain.Duration
}

// GetServerClose returns ServerClose instance
func (deps *ServerDependencies) GetServerClose() util.ServerClose {
	return deps.ServerClose
}

// GetStorage returns Storage instance
func (deps *ServerDependencies) GetStorage() domain.Storage {
	return deps.Storage
}

// GetLongPollingMaxTimeout returns configuration value
func (deps *ServerDependencies) GetLongPollingMaxTimeout() domain.Duration {
	return deps.LongPollingMaxTimeout
}
