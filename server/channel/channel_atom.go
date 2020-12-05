package channel

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
)

// channelAtom is an Channel implementation corresponds to a ChannelConfiguration
type channelAtom struct {
	// ChannelConfig that this object corresponds to.
	config *config.ChannelConfig
}

func newChannelAtom(config *config.ChannelConfig, validate bool) (*channelAtom, error) {
	atom := &channelAtom{
		config: config,
	}
	if validate {
		if err := atom.validate(); err != nil {
			return &channelAtom{}, err
		}
	}
	return atom, nil
}

func (c *channelAtom) validate() error {
	if err := c.validateTemplateStrings(); err != nil {
		return err
	}
	return nil
}

func (c *channelAtom) validateTemplateStrings() error {
	dummy := make(map[string]interface{})
	dummyRegexMatches := make(map[string]string)
	for _, name := range c.config.Regex.GroupNames() {
		dummyRegexMatches[name] = "dummy"
	}
	dummy["regex"] = dummyRegexMatches

	templates := make(map[string]domain.TemplateString)
	for i, webhook := range c.config.Webhooks {
		templates[fmt.Sprintf("webhooks[%d].url", i)] = *webhook.URL
		for name, tpl := range webhook.Headers {
			templates[fmt.Sprintf("webhooks[%d].headers.%s", i, name)] = tpl
		}
	}
	if jwt := c.config.Jwt; jwt != nil {
		for claim, tpl := range jwt.Claims {
			templates[fmt.Sprintf("jwt.claims.%s", claim)] = tpl
		}
	}

	for path, tpl := range templates {
		if _, err := tpl.Execute(dummy); err != nil {
			return xerrors.Errorf("invalid template found on %s: %w", path, err)
		}
	}
	return nil
}

func (c *channelAtom) IsMatch(id domain.ChannelID) bool {
	return c.config.Regex.Match(true, string(id)) != nil
}

func (c *channelAtom) Expire() domain.Duration {
	return *c.config.Expire
}
