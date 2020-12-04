package config

import (
	"fmt"

	"github.com/saiya/dsps/server/domain"
	jwtpkg "github.com/saiya/dsps/server/jwt"
)

var channlesConfigDefault = ChannelsConfig{
	ChannelConfig{
		Regex:  makeRegex(".+"),
		Expire: channelConfigDefaults.Expire,
	},
}
var channelConfigDefaults = ChannelConfig{
	Expire: makeDurationPtr("30m"),
}
var outgoingWebhookConfigDefaults = OutgoingWebhookConfig{
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
}

// ChannelsConfig is list of configured channels
type ChannelsConfig []ChannelConfig

// ChannelConfig represents channel configuration
type ChannelConfig struct {
	Regex  *domain.Regex    `json:"regex"`
	Expire *domain.Duration `json:"expire"`

	Webhooks []OutgoingWebhookConfig `json:"webhooks"`
	Jwt      *JwtValidationConfig    `json:"jwt"`
}

// OutgoingWebhookConfig is webhook configuration of a channel
type OutgoingWebhookConfig struct {
	URL        *domain.TemplateString           `json:"url"`
	Timeout    *domain.Duration                 `json:"timeout"`
	Connection OutgoingWebhookConnectionConfig  `json:"connection"`
	Retry      OutgoingWebhookRetryConfig       `json:"retry"`
	Headers    map[string]domain.TemplateString `json:"headers"`
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

// JwtValidationConfig is JWT configuration of a channel
type JwtValidationConfig struct {
	Alg  string          `json:"alg"`
	Iss  []domain.JwtIss `json:"iss"`
	Keys []string        `json:"keys"`

	Claims map[string]domain.TemplateString `json:"claims"`
}

// PostprocessChannelsConfig fixes/validates config
func PostprocessChannelsConfig(list *ChannelsConfig) error {
	if len(*list) == 0 {
		*list = channlesConfigDefault
	}
	for i := range *list {
		ch := &(*list)[i]
		if err := postprocessChanelConfig(ch); err != nil {
			return fmt.Errorf("error on channels[%d]: %w", i, err)
		}
	}
	return nil
}

func postprocessChanelConfig(ch *ChannelConfig) error {
	if ch.Expire == nil {
		ch.Expire = channelConfigDefaults.Expire
	}
	if err := durationMustBeLargerThanZero("expire", *ch.Expire); err != nil {
		return err
	}

	for i := range ch.Webhooks {
		webhook := &ch.Webhooks[i]
		if err := postprocessWebhookConfig(webhook); err != nil {
			return fmt.Errorf("error on webhooks[%d]: %w", i, err)
		}
	}
	if ch.Jwt != nil {
		if err := postprocessJwtConfig(ch.Jwt); err != nil {
			return fmt.Errorf("error on JWT config: %w", err)
		}
	}
	return nil
}

func postprocessWebhookConfig(webhook *OutgoingWebhookConfig) error {
	if webhook.Timeout == nil {
		webhook.Timeout = outgoingWebhookConfigDefaults.Timeout
	}
	if webhook.Connection.Max == nil {
		webhook.Connection.Max = outgoingWebhookConfigDefaults.Connection.Max
	}
	if webhook.Connection.MaxIdleTime == nil {
		webhook.Connection.MaxIdleTime = outgoingWebhookConfigDefaults.Connection.MaxIdleTime
	}
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
	if webhook.Headers == nil {
		webhook.Headers = make(map[string]domain.TemplateString)
	}

	if err := durationMustBeLargerThanZero("timeout", *webhook.Timeout); err != nil {
		return err
	}
	if err := intMustBeLargerThanZero("connection.max", *webhook.Connection.Max); err != nil {
		return err
	}
	if err := durationMustBeLargerThanZero("connection.maxIdleTime", *webhook.Connection.MaxIdleTime); err != nil {
		return err
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

func postprocessJwtConfig(jwt *JwtValidationConfig) error {
	if jwt.Claims == nil {
		jwt.Claims = make(map[string]domain.TemplateString)
	}

	if len(jwt.Iss) == 0 {
		return fmt.Errorf("must supply one or more \"iss\" (issuer claim) list")
	}
	if jwt.Alg != "none" {
		if len(jwt.Keys) == 0 {
			return fmt.Errorf("must supply one or more \"keys\" to validate JWT signature")
		}
		for i, key := range jwt.Keys {
			if err := jwtpkg.ValidateKey(jwt.Alg, key); err != nil {
				return fmt.Errorf("failed to load keys[%d]: %w", i, err)
			}
		}
	}
	return nil
}
