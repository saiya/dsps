package channel

import (
	"context"
	"fmt"

	"github.com/saiya/dsps/server/domain"
	jwtv "github.com/saiya/dsps/server/jwt/validator"
	"github.com/saiya/dsps/server/webhook/outgoing"
	"golang.org/x/xerrors"
)

type channelImpl struct {
	id    domain.ChannelID
	atoms []*channelAtom

	expire          domain.Duration
	jwtValidators   []jwtv.Validator
	outgoingWebhook outgoing.Client
}

func (c *channelImpl) Expire() domain.Duration {
	return c.expire
}

func newChannelImpl(id domain.ChannelID, atoms []*channelAtom) (*channelImpl, error) {
	expire := domain.Duration{Duration: 0}
	jwtValidators := make([]jwtv.Validator, 0, len(atoms))
	outgoingWebhooks := make([]outgoing.Client, 0, len(atoms)*2)
	for _, atom := range atoms {
		tplEnv := atom.TemplateEnvironmentOf(id)
		if tplEnv == nil {
			return nil, fmt.Errorf(`failed to evaluate channel configuration /%s/ to channel "%s"`, atom.String(), id)
		}

		if expire.Duration < atom.Expire().Duration {
			expire = atom.Expire()
		}

		if atom.JwtValidatorTemplate != nil {
			jv, err := atom.JwtValidatorTemplate.NewValidator(tplEnv)
			if err != nil {
				return nil, xerrors.Errorf(`failed to configure JWT validation of channel "%s": %w`, id, err)
			}
			jwtValidators = append(jwtValidators, jv)
		}

		for _, tpl := range atom.OutgoingWebHookTemplates {
			client, err := tpl.NewClient(tplEnv)
			if err != nil {
				return nil, xerrors.Errorf(`failed to setup outgoing webhook of channel "%s": %w`, id, err)
			}
			outgoingWebhooks = append(outgoingWebhooks, client)
		}
	}
	return &channelImpl{
		id:    id,
		atoms: atoms,

		expire:          expire,
		jwtValidators:   jwtValidators,
		outgoingWebhook: outgoing.NewMultiplexClient(outgoingWebhooks),
	}, nil
}

func (c *channelImpl) ValidateJwt(ctx context.Context, jwt string) error {
	for _, jv := range c.jwtValidators {
		if err := jv.Validate(ctx, jwt); err != nil {
			return err
		}
	}
	return nil
}

func (c *channelImpl) SendOutgoingWebhook(ctx context.Context, msg domain.Message) error {
	return c.outgoingWebhook.Send(ctx, msg)
}
