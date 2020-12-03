package config

import (
	"regexp"
	"strings"

	"github.com/saiya/dsps/server/domain"
)

// HTTPServerConfig represents HTTP webserver settings
type HTTPServerConfig struct {
	Port                int    `json:"port" validate:"min=0,max=65535"`
	PathPrefix          string `json:"pathPrefix"`
	SourceIPHeader      string `json:"sourceIpHeader"`
	ShowForbiddenDetail bool   `json:"showForbiddenDetail"`

	LongPollingMaxTimeout   domain.Duration `json:"longPollingMaxTimeout"`
	GracefulShutdownTimeout domain.Duration `json:"gracefulShutdownTimeout"`
}

var httpServerConfigDefault = HTTPServerConfig{
	Port:                3000,
	SourceIPHeader:      "",
	ShowForbiddenDetail: false,

	LongPollingMaxTimeout:   makeDuration("30s"),
	GracefulShutdownTimeout: makeDuration("5s"),
}

// PostprocessHTTPServerConfig cleanups user supplied config object.
func PostprocessHTTPServerConfig(config *HTTPServerConfig, overrides Overrides) error {
	if overrides.Port != 0 {
		config.Port = overrides.Port
	}

	// Remove trailing "/", add "/" prefix
	config.PathPrefix = regexp.MustCompile(`/$`).ReplaceAllString(config.PathPrefix, "")
	if !strings.HasPrefix(config.PathPrefix, "/") {
		config.PathPrefix = "/" + config.PathPrefix
	}

	return nil
}
