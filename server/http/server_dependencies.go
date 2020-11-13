package http

import (
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
)

// ServerDependencies struct holds all resource references to build web server
type ServerDependencies struct {
	Logger          logger.Logger
	DebugLogEnabler logger.DebugLogEnabler

	Storage domain.Storage
}

// GetStorage returns Storage instance
func (deps *ServerDependencies) GetStorage() domain.Storage {
	return deps.Storage
}
