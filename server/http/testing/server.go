package testing

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/channel"
	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http"
	httplifecycle "github.com/saiya/dsps/server/http/lifecycle"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/storage"
)

// WithServerDeps runs given test function with ServerDependencies
func WithServerDeps(t *testing.T, configYaml string, f func(*http.ServerDependencies)) {
	cfg, err := config.ParseConfig(config.Overrides{}, strings.ReplaceAll(configYaml, "\t", "  "))
	if !assert.NoError(t, err) {
		return
	}

	assert.NoError(t, logger.InitLogger(cfg.Logging))

	ctx := context.Background()
	clock := domain.RealSystemClock
	channelProvider, err := channel.NewChannelProvider(ctx, &cfg, clock)
	assert.NoError(t, err)
	storage, err := storage.NewStorage(ctx, &cfg.Storages, clock, channelProvider)
	assert.NoError(t, err)
	serverClose := httplifecycle.NewServerClose()
	defer serverClose.Close()

	f(&http.ServerDependencies{
		Config:          &cfg,
		ChannelProvider: channelProvider,
		Storage:         storage,
		ServerClose:     serverClose,
	})
}

// WithServer runs given test function with HTTP server
func WithServer(t *testing.T, configYaml string, setup func(deps *http.ServerDependencies), f func(deps *http.ServerDependencies, baseURL string)) {
	WithServerDeps(t, configYaml, func(deps *http.ServerDependencies) {
		setup(deps)

		server := http.CreateServer(context.Background(), deps)
		ts := httptest.NewServer(server)
		defer ts.Close()
		defer deps.ServerClose.Close() // Close requests first

		f(deps, strings.TrimSuffix(ts.URL, "/")+strings.TrimSuffix(deps.Config.HTTPServer.PathPrefix, "/"))
	})
}
