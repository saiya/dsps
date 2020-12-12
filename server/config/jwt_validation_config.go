package config

import (
	"fmt"

	"github.com/saiya/dsps/server/domain"
	jwtpkg "github.com/saiya/dsps/server/jwt"
)

// JwtValidationConfig is JWT configuration of a channel
type JwtValidationConfig struct {
	Alg             domain.JwtAlg    `json:"alg"`
	ClockSkewLeeway *domain.Duration `json:"clockSkewLeeway"`

	Iss  []domain.JwtIss `json:"iss"`
	Aud  []domain.JwtAud `json:"aud"`
	Keys []string        `json:"keys"`

	Claims map[string]domain.TemplateString `json:"claims"`
}

func postprocessJwtConfig(jwt *JwtValidationConfig) error {
	if jwt.Claims == nil {
		jwt.Claims = make(map[string]domain.TemplateString)
	}
	if jwt.ClockSkewLeeway == nil {
		jwt.ClockSkewLeeway = makeDurationPtr("5m")
	}

	if err := jwtpkg.ValidateAlg(jwt.Alg); err != nil {
		return fmt.Errorf("invalid \"alg\": %w", err)
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
