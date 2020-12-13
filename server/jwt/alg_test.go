package jwt_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/jwt"
)

func TestValidateAlg(t *testing.T) {
	for _, supported := range []string{
		// ref: https://tools.ietf.org/html/rfc7518#section-3.1
		"HS256", "HS384", "HS512", // HMAC
		"RS256", "RS384", "RS512", // RSASSA-PKCS1
		"ES256", "ES384", "ES512", // ECDSA
		"PS256", "PS384", "PS512", // RSASSA-PSS
		"none",
	} {
		assert.NoError(t, ValidateAlg(domain.JwtAlg(supported)))
	}

	for _, invalid := range []string{
		"A256KW", "RSA-OAEP", "dir", "ECDH-ES", "ECDH-ES+A256KW", "A256GCMKW", "PBES2-HS512+A256KW", // Those are JWE alg
	} {
		assert.Equal(t, fmt.Sprintf(`Unsupported JWT alg "%s"`, invalid), ValidateAlg(domain.JwtAlg(invalid)).Error())
	}
}

func TestAlgTypeDetection(t *testing.T) {
	assert.True(t, IsRSA(domain.JwtAlg("RS512"))) // RSASSA-PKCS1
	assert.True(t, IsRSA(domain.JwtAlg("PS512"))) // RSASSA-PSS
	assert.True(t, IsECDSA(domain.JwtAlg("ES512")))
	assert.True(t, IsHMAC(domain.JwtAlg("HS512")))
}
