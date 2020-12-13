package config

import (
	"fmt"

	"github.com/saiya/dsps/server/domain"
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

// ChannelsConfig is list of configured channels
type ChannelsConfig []ChannelConfig

// ChannelConfig represents channel configuration
type ChannelConfig struct {
	Regex  *domain.Regex    `json:"regex"`
	Expire *domain.Duration `json:"expire"`

	Webhooks []OutgoingWebhookConfig `json:"webhooks"`
	Jwt      *JwtValidationConfig    `json:"jwt"`
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
