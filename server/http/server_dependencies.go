package http

import (
	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/lifecycle"
	"github.com/saiya/dsps/server/logger"
)

// ServerDependencies struct holds all resource references to build web server
type ServerDependencies struct {
	Config          *config.ServerConfig
	ChannelProvider domain.ChannelProvider
	Storage         domain.Storage

	LogFilter   *logger.Filter
	ServerClose lifecycle.ServerClose
}

// GetChannelProvider returns ChannelProvider object
func (deps *ServerDependencies) GetChannelProvider() domain.ChannelProvider {
	return deps.ChannelProvider
}

// GetStorage returns Storage instance
func (deps *ServerDependencies) GetStorage() domain.Storage {
	return deps.Storage
}

// GetDefaultHeaders returns default response headers config
func (deps *ServerDependencies) GetDefaultHeaders() map[string]string {
	return deps.Config.HTTPServer.DefaultHeaders
}

// GetLongPollingMaxTimeout returns configuration value
func (deps *ServerDependencies) GetLongPollingMaxTimeout() domain.Duration {
	return deps.Config.HTTPServer.LongPollingMaxTimeout
}

// DiscloseAuthRejectionDetail returns configuration value
func (deps *ServerDependencies) DiscloseAuthRejectionDetail() bool {
	return deps.Config.HTTPServer.DiscloseAuthRejectionDetail
}

// GetIPHeaderName returns configuration value
func (deps *ServerDependencies) GetIPHeaderName() string {
	return deps.Config.HTTPServer.RealIPHeader
}

// GetTrustedProxyRanges returns configuration value
func (deps *ServerDependencies) GetTrustedProxyRanges() []domain.CIDR {
	return deps.Config.HTTPServer.TrustedProxyRanges
}

// GetAdminAuthConfig returns configuration value
func (deps *ServerDependencies) GetAdminAuthConfig() *config.AdminAuthConfig {
	return &deps.Config.Admin.Auth
}

// GetLogFilter returns log filter instance
func (deps *ServerDependencies) GetLogFilter() *logger.Filter {
	return deps.LogFilter
}

// GetServerClose returns ServerClose instance
func (deps *ServerDependencies) GetServerClose() lifecycle.ServerClose {
	return deps.ServerClose
}
