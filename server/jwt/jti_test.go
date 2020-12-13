package jwt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/jwt"
	. "github.com/saiya/dsps/server/jwt/testing"
)

func TestExtractJti(t *testing.T) {
	jti, err := ExtractJti(GenerateJwt(t, JwtProps{
		Alg:     "ES512",
		Keyname: "ES512-test1",
		JwtDir:  ".",
		Jti:     "86853062-128F-45D9-99CA-D5E7585C9A6C",
	}))
	assert.NoError(t, err)
	assert.Equal(t, domain.JwtJti("86853062-128F-45D9-99CA-D5E7585C9A6C"), *jti)

	jti, err = ExtractJti(GenerateJwt(t, JwtProps{
		Alg:     "ES512",
		Keyname: "ES512-test1",
		JwtDir:  ".",
	}))
	assert.NoError(t, err)
	assert.Nil(t, jti)

	_, err = ExtractJti(`this-is-not-JWT`)
	assert.Error(t, err)
	assert.Regexp(t, `token is malformed: token contains an invalid number of segments`, err.Error())
}
