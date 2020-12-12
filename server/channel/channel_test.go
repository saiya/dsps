package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/channel"
	. "github.com/saiya/dsps/server/jwt/testing"
)

func TestChannelExpire(t *testing.T) {
	assert.Equal(t, 35*time.Minute, channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m' }`,
	}).Expire().Duration)
	assert.Equal(t, 105*time.Minute, channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m' }`,
		`{ regex: '.+', expire: '105m' }`,
	}).Expire().Duration)
}

func TestJwtValidation(t *testing.T) {
	ctx := context.Background()

	// No JWT validation configured.
	assert.NoError(t, channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m' }`,
	}).ValidateJwt(ctx, ""))
	assert.NoError(t, channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m' }`,
	}).ValidateJwt(ctx, "this is not JWT"))

	// Malformed JWT
	err := channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m', jwt: { iss: [ "https://example.com/issuer" ], keys: { ES512: [ "../jwt/testdata/ES512-test1-public.pem" ] } } }`,
	}).ValidateJwt(ctx, "")
	assert.Error(t, err)
	assert.Regexp(t, "no JWT presented", err.Error())
	err = channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m', jwt: { iss: [ "https://example.com/issuer" ], keys: { ES512: [ "../jwt/testdata/ES512-test1-public.pem" ] } } }`,
	}).ValidateJwt(ctx, "this is not JWT")
	assert.Error(t, err)
	assert.Regexp(t, "JWT validation failed: token is malformed", err.Error())

	// Valid JWT
	assert.NoError(t, channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m', jwt: { iss: [ "https://example.com/issuer" ], keys: { ES512: [ "../jwt/testdata/ES512-test1-public.pem" ] } } }`,
	}).ValidateJwt(ctx, GenerateJwt(t, JwtProps{
		Alg:     "ES512",
		JwtDir:  "../jwt",
		Keyname: "ES512-test1",
		Iss:     "https://example.com/issuer",
	})))

	// Multiple atom builds AND condition
	multiValidation := channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m', jwt: { iss: [ "https://example.com/issuer", "https://example.com/issuer2" ], keys: { ES512: [ "../jwt/testdata/ES512-test1-public.pem" ] } } }`,
		`{ regex: '.+', expire: '35m', jwt: { iss: [ "https://example.com/issuer", "https://example.com/issuer3" ], keys: { ES512: [ "../jwt/testdata/ES512-test1-public.pem" ] } } }`,
	})
	assert.NoError(t, multiValidation.ValidateJwt(ctx, GenerateJwt(t, JwtProps{
		Alg:     "ES512",
		JwtDir:  "../jwt",
		Keyname: "ES512-test1",
		Iss:     "https://example.com/issuer",
	})))
	err = multiValidation.ValidateJwt(ctx, GenerateJwt(t, JwtProps{
		Alg:     "ES512",
		JwtDir:  "../jwt",
		Keyname: "ES512-test1",
		Iss:     "https://example.com/issuer2", // Unmatch with one atom
	}))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `"iss" claim of the presented JWT ("https://example.com/issuer2") does not match with any of expected values`)
}
