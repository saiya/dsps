package testing

import (
	"testing"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go/v4"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/jwt"
	"github.com/stretchr/testify/assert"
)

// JwtProps is only for GenerateJwt
type JwtProps struct {
	// Filename prefix under ../testdata (e.g. "ES512-test1")
	Keyname string
	// Path to /server/jwt dir
	JwtDir string

	Alg domain.JwtAlg
	// Issuer
	Iss domain.JwtIss
	// Issue at
	Iat time.Time
	// Expire at
	Exp time.Time
	// Not valid before
	Nbf time.Time
	Aud []domain.JwtAud

	Claims map[string]interface{}
}

// GenerateJwt generates JWT only for testing purpose.
func GenerateJwt(t *testing.T, props JwtProps) string {
	claims := jwtgo.MapClaims{"alg": string(props.Alg)}
	if props.Iss != "" {
		claims["iss"] = props.Iss
	}
	if props.Iat != (time.Time{}) {
		claims["iat"] = props.Iat.Unix()
	}
	if props.Exp != (time.Time{}) {
		claims["exp"] = props.Exp.Unix()
	}
	if props.Nbf != (time.Time{}) {
		claims["nbf"] = props.Nbf.Unix()
	}
	switch len(props.Aud) {
	case 0:
	case 1:
		claims["aud"] = string(props.Aud[0])
	default:
		strs := make([]string, len(props.Aud))
		for i, aud := range props.Aud {
			strs[i] = string(aud)
		}
		claims["aud"] = strs
	}
	for key, value := range props.Claims {
		claims[key] = value
	}

	token := jwtgo.NewWithClaims(jwtgo.GetSigningMethod(string(props.Alg)), claims)
	if props.Alg == "none" {
		jwt, err := token.SigningString()
		assert.NoError(t, err)
		return jwt + "."
	}
	if props.JwtDir == "" {
		props.JwtDir = ".."
	}
	key, err := jwt.LoadKey(props.Alg, props.JwtDir+"/testdata/"+props.Keyname+"-private.pem", true)
	assert.NoError(t, err)
	jwt, err := token.SignedString(key)
	assert.NoError(t, err)
	// fmt.Printf("JWT generated: %s\n", jwt)
	return jwt
}
