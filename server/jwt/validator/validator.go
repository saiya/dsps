package validator

import (
	"context"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
)

// Validator is a object to validate JWT
type Validator interface {
	Validate(ctx context.Context, jwt string, templateStringEnv interface{}) error
}

// NewValidator creates validator instance.
func NewValidator(ctx context.Context, cfg *config.JwtValidationConfig, clock domain.SystemClock) (Validator, error) {
	return &validator{cfg: cfg, clock: clock}, nil
}

type validator struct {
	cfg   *config.JwtValidationConfig
	clock domain.SystemClock
}

func (v *validator) Validate(ctx context.Context, jwt string, templateStringEnv interface{}) error {
	return nil // FIXME: Impls
}
