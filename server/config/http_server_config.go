package config

import (
	"fmt"
	"strings"

	"github.com/saiya/dsps/server/domain"
)

// HTTPServerConfig represents HTTP webserver settings
type HTTPServerConfig struct {
	Port                        int    `json:"port" validate:"min=0,max=65535"`
	Listen                      string `json:"listen"`
	PathPrefix                  string `json:"pathPrefix"`
	SourceIPHeader              string `json:"sourceIpHeader"`
	DiscloseAuthRejectionDetail bool   `json:"discloseAuthRejectionDetail"`

	LongPollingMaxTimeout   domain.Duration `json:"longPollingMaxTimeout"`
	GracefulShutdownTimeout domain.Duration `json:"gracefulShutdownTimeout"`
}

var httpServerConfigDefault = HTTPServerConfig{
	Port:                        3000,
	Listen:                      "",
	SourceIPHeader:              "",
	DiscloseAuthRejectionDetail: false,

	LongPollingMaxTimeout:   makeDuration("30s"),
	GracefulShutdownTimeout: makeDuration("5s"),
}

// PostprocessHTTPServerConfig cleanups user supplied config object.
func PostprocessHTTPServerConfig(config *HTTPServerConfig, overrides Overrides) error {
	if overrides.Port != 0 {
		config.Port = overrides.Port
	}
	if config.Listen == "" {
		config.Listen = fmt.Sprintf(":%d", config.Port)
	}
	if overrides.Listen != "" {
		config.Listen = overrides.Listen
	}

	// Remove "/" prefix and suffix
	config.PathPrefix = strings.TrimPrefix(strings.TrimSuffix(config.PathPrefix, "/"), "/")
	return nil
}
