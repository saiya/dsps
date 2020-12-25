package channel

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	jwtv "github.com/saiya/dsps/server/jwt/validator"
	"github.com/saiya/dsps/server/webhook/outgoing"
)

// channelAtom is an Channel implementation corresponds to a ChannelConfiguration
type channelAtom struct {
	// ChannelConfig that this object corresponds to.
	config *config.ChannelConfig

	JwtValidatorTemplate     jwtv.Template
	OutgoingWebHookTemplates []outgoing.ClientTemplate
}

func newChannelAtom(ctx context.Context, config *config.ChannelConfig, deps ProviderDeps, validate bool) (*channelAtom, error) {
	if err := deps.validateProviderDeps(); err != nil {
		return nil, err
	}

	atom := &channelAtom{
		config: config,
	}
	if validate {
		if err := atom.validate(); err != nil {
			return nil, err
		}
	}

	if config.Jwt != nil {
		jvt, err := jwtv.NewTemplate(ctx, config.Jwt, deps.Clock)
		if err != nil {
			return nil, err
		}
		atom.JwtValidatorTemplate = jvt
	}

	atom.OutgoingWebHookTemplates = make([]outgoing.ClientTemplate, 0, len(config.Webhooks))
	for i := range config.Webhooks {
		tpl, err := outgoing.NewClientTemplate(ctx, &config.Webhooks[i], deps.Telemetry, deps.Sentry)
		if err != nil {
			return nil, err
		}
		atom.OutgoingWebHookTemplates = append(atom.OutgoingWebHookTemplates, tpl)
	}

	return atom, nil
}

func (c *channelAtom) Shutdown(ctx context.Context) {
	for _, webhook := range c.OutgoingWebHookTemplates {
		webhook.Close()
	}
}

func (c *channelAtom) String() string {
	return c.config.Regex.String()
}

func (c *channelAtom) GetFileDescriptorPressure() int {
	result := 0
	for _, webhook := range c.OutgoingWebHookTemplates {
		result += webhook.GetFileDescriptorPressure()
	}
	return result
}

func (c *channelAtom) validate() error {
	if err := c.validateTemplateStrings(); err != nil {
		return err
	}
	return nil
}

func (c *channelAtom) validateTemplateStrings() error {
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

	dummy := c.dummyTemplateEnvironment()
	for path, tpl := range templates {
		if _, err := tpl.Execute(dummy); err != nil {
			return xerrors.Errorf("invalid template found on %s: %w", path, err)
		}
	}
	return nil
}

func (c *channelAtom) TemplateEnvironmentOf(id domain.ChannelID) domain.TemplateStringEnv {
	matches := c.config.Regex.Match(true, string(id))
	if matches == nil {
		return nil
	}
	return map[string]interface{}{
		"channel": matches,
	}
}

func (c *channelAtom) dummyTemplateEnvironment() domain.TemplateStringEnv {
	dummyRegexMatches := make(map[string]string)
	for _, name := range c.config.Regex.GroupNames() {
		dummyRegexMatches[name] = "dummy"
	}
	return map[string]interface{}{
		"channel": dummyRegexMatches,
	}
}

func (c *channelAtom) IsMatch(id domain.ChannelID) bool {
	return c.config.Regex.Match(true, string(id)) != nil
}

func (c *channelAtom) Expire() domain.Duration {
	return *c.config.Expire
}
