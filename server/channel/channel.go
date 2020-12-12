package channel

import (
	"context"
	"fmt"

	"github.com/saiya/dsps/server/domain"
)

type channelImpl struct {
	id    domain.ChannelID
	atoms []struct {
		*channelAtom
		tplEnv domain.TemplateStringEnv
	}

	expire domain.Duration
}

func (c *channelImpl) Expire() domain.Duration {
	return c.expire
}

func newChannelImpl(id domain.ChannelID, atoms []*channelAtom) *channelImpl {
	mappedAtoms := make([]struct {
		*channelAtom
		tplEnv domain.TemplateStringEnv
	}, len(atoms))
	for i, atom := range atoms {
		tplEnv := atom.TemplateEnvironmentOf(id)
		if tplEnv == nil {
			panic(fmt.Errorf(`Failed to evaluate channel configuration /%s/ to channel "%s"`, atom.String(), id))
		}
		mappedAtoms[i] = struct {
			*channelAtom
			tplEnv domain.TemplateStringEnv
		}{channelAtom: atom, tplEnv: tplEnv}
	}

	c := &channelImpl{
		id:    id,
		atoms: mappedAtoms,

		expire: atoms[0].Expire(),
	}
	for _, atom := range atoms {
		if c.expire.Duration < atom.Expire().Duration {
			c.expire = atom.Expire()
		}
	}
	return c
}

func (c *channelImpl) ValidateJwt(ctx context.Context, jwt string) error {
	for _, atom := range c.atoms {
		if atom.JwtValidator != nil {
			if err := atom.JwtValidator.Validate(ctx, jwt, atom.tplEnv); err != nil {
				return err
			}
		}
	}
	return nil
}
