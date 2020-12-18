package config

import (
	"fmt"
	"strings"

	"github.com/saiya/dsps/server/domain"
)

// OutgoingWebhookConfig is webhook configuration of a channel
type OutgoingWebhookConfig struct {
	Method     string                           `json:"method"`
	URL        *domain.TemplateString           `json:"url"`
	Timeout    *domain.Duration                 `json:"timeout"`
	Connection OutgoingWebhookConnectionConfig  `json:"connection"`
	Retry      OutgoingWebhookRetryConfig       `json:"retry"`
	Headers    map[string]domain.TemplateString `json:"headers"`

	MaxRedirects *int `json:"maxRedirects"`
}

var validWebhookMethods = map[string]interface{}{
	"PUT":  struct{}{},
	"POST": struct{}{},
}

// OutgoingWebhookConnectionConfig is HTTP/TCP connection config
type OutgoingWebhookConnectionConfig struct {
	Max         *int             `json:"max"`
	MaxIdleTime *domain.Duration `json:"maxIdleTime"`
}

// OutgoingWebhookRetryConfig is retry config
type OutgoingWebhookRetryConfig struct {
	Count              *int             `json:"count"`
	Interval           *domain.Duration `json:"interval"`
	IntervalMultiplier *float64         `json:"intervalMultiplier"`
	IntervalJitter     *domain.Duration `json:"intervalJitter"`
}

var outgoingWebhookConfigDefaults = OutgoingWebhookConfig{
	Method:  "PUT",
	Timeout: makeDurationPtr("30s"),
	Connection: OutgoingWebhookConnectionConfig{
		Max:         makeIntPtr(1024),
		MaxIdleTime: makeDurationPtr("3m"),
	},
	Retry: OutgoingWebhookRetryConfig{
		Count:              makeIntPtr(3),
		Interval:           makeDurationPtr("3s"),
		IntervalMultiplier: makeFloat64Ptr(1.5),
		IntervalJitter:     makeDurationPtr("1s500ms"),
	},
	MaxRedirects: makeIntPtr(10),
}

func postprocessWebhookConfig(webhook *OutgoingWebhookConfig) error {
	if webhook.Method == "" {
		webhook.Method = outgoingWebhookConfigDefaults.Method
	}
	webhook.Method = strings.ToUpper(webhook.Method)
	if webhook.Timeout == nil {
		webhook.Timeout = outgoingWebhookConfigDefaults.Timeout
	}
	if webhook.Headers == nil {
		webhook.Headers = make(map[string]domain.TemplateString)
	}
	if webhook.MaxRedirects == nil {
		webhook.MaxRedirects = outgoingWebhookConfigDefaults.MaxRedirects
	}

	if err := postprocessWebhookRetryConfig(webhook); err != nil {
		return err
	}
	if err := postprocessWebhookConnectionConfig(webhook); err != nil {
		return err
	}

	if _, ok := validWebhookMethods[webhook.Method]; !ok {
		return fmt.Errorf(`"%s" is not valid outgoing-webhook HTTP method`, webhook.Method)
	}
	if err := durationMustBeLargerThanZero("timeout", *webhook.Timeout); err != nil {
		return err
	}
	if *webhook.MaxRedirects < 0 {
		return fmt.Errorf("maxRedirects must not be negative: %d", *webhook.MaxRedirects)
	}
	return nil
}

func postprocessWebhookRetryConfig(webhook *OutgoingWebhookConfig) error {
	if webhook.Retry.Count == nil {
		webhook.Retry.Count = outgoingWebhookConfigDefaults.Retry.Count
	}
	if webhook.Retry.Interval == nil {
		webhook.Retry.Interval = outgoingWebhookConfigDefaults.Retry.Interval
	}
	if webhook.Retry.IntervalMultiplier == nil {
		webhook.Retry.IntervalMultiplier = outgoingWebhookConfigDefaults.Retry.IntervalMultiplier
	}
	if webhook.Retry.IntervalJitter == nil {
		webhook.Retry.IntervalJitter = outgoingWebhookConfigDefaults.Retry.IntervalJitter
	}

	if err := intMustBeLargerThanZero("retry.count", *webhook.Retry.Count); err != nil {
		return err
	}
	if err := durationMustBeLargerThanZero("retry.interval", *webhook.Retry.Interval); err != nil {
		return err
	}
	if *webhook.Retry.IntervalMultiplier < 1.0 {
		return fmt.Errorf("retry.intervalMultipler must be equal to or larger than 1.0")
	}
	if err := durationMustBeLargerThanZero("retry.intervalJitter", *webhook.Retry.IntervalJitter); err != nil {
		return err
	}
	return nil
}

func postprocessWebhookConnectionConfig(webhook *OutgoingWebhookConfig) error {
	if webhook.Connection.Max == nil {
		webhook.Connection.Max = outgoingWebhookConfigDefaults.Connection.Max
	}
	if webhook.Connection.MaxIdleTime == nil {
		webhook.Connection.MaxIdleTime = outgoingWebhookConfigDefaults.Connection.MaxIdleTime
	}

	if err := intMustBeLargerThanZero("connection.max", *webhook.Connection.Max); err != nil {
		return err
	}
	if err := durationMustBeLargerThanZero("connection.maxIdleTime", *webhook.Connection.MaxIdleTime); err != nil {
		return err
	}

	return nil
}
