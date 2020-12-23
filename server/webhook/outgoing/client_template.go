package outgoing

import (
	"context"
	"net/http"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/telemetry"
)

// ClientTemplate is factory object to make Client
type ClientTemplate interface {
	NewClient(tplEnv domain.TemplateStringEnv) (Client, error)
	Close()

	GetFileDescriptorPressure() int // estimated max usage of file descriptors
}

type clientTemplate struct {
	*config.OutgoingWebhookConfig

	h        *http.Client
	maxConns int

	telemetry *telemetry.Telemetry
}

// NewClientTemplate returns ClientTemplate instalce
func NewClientTemplate(ctx context.Context, cfg *config.OutgoingWebhookConfig, telemetry *telemetry.Telemetry) (ClientTemplate, error) {
	return &clientTemplate{
		OutgoingWebhookConfig: cfg,

		h:        newHTTPClientFor(ctx, cfg),
		maxConns: *cfg.Connection.Max,

		telemetry: telemetry,
	}, nil
}

func (tpl *clientTemplate) NewClient(tplEnv domain.TemplateStringEnv) (Client, error) {
	return newClientImpl(tpl, tplEnv)
}

func (tpl *clientTemplate) Close() {
	tpl.h.CloseIdleConnections()
}

func (tpl *clientTemplate) GetFileDescriptorPressure() int {
	return tpl.maxConns
}
