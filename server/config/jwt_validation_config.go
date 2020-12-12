package config

import (
	"fmt"

	"github.com/saiya/dsps/server/domain"
	jwtpkg "github.com/saiya/dsps/server/jwt"
)

// JwtValidationConfig is JWT configuration of a channel
type JwtValidationConfig struct {
	Iss  []domain.JwtIss            `json:"iss"`
	Aud  []domain.JwtAud            `json:"aud"`
	Keys map[domain.JwtAlg][]string `json:"keys"`

	Claims map[string]domain.TemplateString `json:"claims"`

	ClockSkewLeeway *domain.Duration `json:"clockSkewLeeway"`
}

func postprocessJwtConfig(jwt *JwtValidationConfig) error {
	if jwt.Claims == nil {
		jwt.Claims = make(map[string]domain.TemplateString)
	}
	if jwt.ClockSkewLeeway == nil {
		jwt.ClockSkewLeeway = makeDurationPtr("5m")
	}

	if len(jwt.Iss) == 0 {
		return fmt.Errorf(`must supply one or more "iss" (issuer claim) list`)
	}

	if len(jwt.Keys) == 0 {
		return fmt.Errorf(`must supply one or more "keys" (signing algorithm and keys) setting`)
	}
	for alg, keyFiles := range jwt.Keys {
		if err := jwtpkg.ValidateAlg(alg); err != nil {
			return fmt.Errorf(`invalid signing algorithm name given "%s": %w`, alg, err)
		}
		if !alg.IsNone() {
			if len(keyFiles) == 0 {
				return fmt.Errorf("must supply one or more key file(s) to validate JWT signature for alg=%s", alg)
			}
			for i, keyFile := range keyFiles {
				if err := jwtpkg.ValidateVerificationKey(alg, keyFile); err != nil {
					return fmt.Errorf("failed to load keys[%s][%d]: %w", alg, i, err)
				}
			}
		}
	}
	return nil
}
