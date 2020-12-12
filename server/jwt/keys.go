package jwt

import (
	"github.com/saiya/dsps/server/domain"
)

// ValidateKey validates plain-text representation of public key
func ValidateKey(alg domain.JwtAlg, key string) error {
	// TODO: Implement
	return nil
}
