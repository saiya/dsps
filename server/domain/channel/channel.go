package channel

import (
	"context"
	"fmt"

	"github.com/saiya/dsps/server/domain"
	jwtv "github.com/saiya/dsps/server/jwt/validator"
	"golang.org/x/xerrors"
)

type channelImpl struct {
	id    domain.ChannelID
	atoms []*channelAtom

	expire        domain.Duration
	jwtValidators []jwtv.Validator
}

func (c *channelImpl) Expire() domain.Duration {
	return c.expire
}

func newChannelImpl(id domain.ChannelID, atoms []*channelAtom) (*channelImpl, error) {
	jwtValidators := make([]jwtv.Validator, 0, len(atoms))
	for _, atom := range atoms {
		tplEnv := atom.TemplateEnvironmentOf(id)
		if tplEnv == nil {
			return nil, fmt.Errorf(`failed to evaluate channel configuration /%s/ to channel "%s"`, atom.String(), id)
		}
		if atom.JwtValidatorTemplate != nil {
			jv, err := atom.JwtValidatorTemplate.NewValidator(tplEnv)
			if err != nil {
				return nil, xerrors.Errorf(`failed to configure JWT validation of channel "%s": %w`, id, err)
			}
			jwtValidators = append(jwtValidators, jv)
		}
	}

	c := &channelImpl{
		id:    id,
		atoms: atoms,

		expire:        atoms[0].Expire(),
		jwtValidators: jwtValidators,
	}
	for _, atom := range atoms {
		if c.expire.Duration < atom.Expire().Duration {
			c.expire = atom.Expire()
		}
	}
	return c, nil
}

func (c *channelImpl) ValidateJwt(ctx context.Context, jwt string) error {
	for _, jv := range c.jwtValidators {
		if err := jv.Validate(ctx, jwt); err != nil {
			return err
		}
	}
	return nil
}
