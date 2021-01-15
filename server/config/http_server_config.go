package config

import (
	"fmt"
	"strings"

	"github.com/saiya/dsps/server/domain"
)

// HTTPServerConfig represents HTTP webserver settings
type HTTPServerConfig struct {
	Port                        int               `json:"port" validate:"min=0,max=65535"`
	Listen                      string            `json:"listen"`
	PathPrefix                  string            `json:"pathPrefix"`
	RealIPHeader                string            `json:"realIpHeader"`
	TrustedProxyRanges          []domain.CIDR     `json:"trustedProxyRanges"`
	DiscloseAuthRejectionDetail bool              `json:"discloseAuthRejectionDetail"`
	DefaultHeaders              map[string]string `json:"defaultHeaders"`

	IdleTimeout             domain.Duration `json:"idleTimeout"`
	ReadTimeout             domain.Duration `json:"readTimeout"`
	WriteTimeout            domain.Duration `json:"writeTimeout"`
	LongPollingMaxTimeout   domain.Duration `json:"longPollingMaxTimeout"`
	GracefulShutdownTimeout domain.Duration `json:"gracefulShutdownTimeout"`
}

func httpServerConfigDefault() *HTTPServerConfig {
	return &HTTPServerConfig{
		Port: 3000,

		IdleTimeout:             makeDuration("1h30m"),
		ReadTimeout:             makeDuration("10s"),
		WriteTimeout:            makeDuration("60s"),
		LongPollingMaxTimeout:   makeDuration("30s"),
		GracefulShutdownTimeout: makeDuration("5s"),
	}
}

var defaultHeaders = map[string]string{
	"X-Frame-Options":        `deny`,
	"X-Content-Type-Options": `nosniff`,
	"Cache-Control":          `no-cache, no-store, max-age=0, must-revalidate`,
	"Pragma":                 `no-cache`,
	"Expires":                `0`,
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

	if config.Port == 0 {
		config.Port = httpServerConfigDefault().Port
	}
	if len(config.TrustedProxyRanges) == 0 {
		config.TrustedProxyRanges = make([]domain.CIDR, len(domain.PrivateCIDRs))
		copy(config.TrustedProxyRanges, domain.PrivateCIDRs)
	}
	if config.IdleTimeout.Duration == 0 {
		config.IdleTimeout = httpServerConfigDefault().IdleTimeout
	}
	if config.ReadTimeout.Duration == 0 {
		config.ReadTimeout = httpServerConfigDefault().ReadTimeout
	}
	if config.WriteTimeout.Duration == 0 {
		config.WriteTimeout = httpServerConfigDefault().WriteTimeout
	}
	if config.LongPollingMaxTimeout.Duration == 0 {
		config.LongPollingMaxTimeout = httpServerConfigDefault().LongPollingMaxTimeout
	}
	if config.GracefulShutdownTimeout.Duration == 0 {
		config.GracefulShutdownTimeout = httpServerConfigDefault().GracefulShutdownTimeout
	}

	if config.DefaultHeaders == nil {
		config.DefaultHeaders = make(map[string]string, len(defaultHeaders))
	}
	for name, value := range defaultHeaders {
		if _, ok := config.DefaultHeaders[name]; !ok {
			config.DefaultHeaders[name] = value
		}
	}

	// Remove "/" prefix and suffix
	config.PathPrefix = strings.TrimPrefix(strings.TrimSuffix(config.PathPrefix, "/"), "/")
	return nil
}
