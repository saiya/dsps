package jwt

import (
	jwtgo "github.com/dgrijalva/jwt-go/v4"

	"github.com/saiya/dsps/server/domain"
)

// ExtractJti read "jti" claim of JWT. Does not perform any JWT validation.
func ExtractJti(jwtStr string) (*domain.JwtJti, error) {
	parser := jwtgo.NewParser()
	claims := jwtgo.StandardClaims{}
	if _, _, err := parser.ParseUnverified(jwtStr, &claims); err != nil {
		return nil, err
	}
	if claims.ID != "" {
		jti := domain.JwtJti(claims.ID)
		return &jti, nil
	}
	return nil, nil
}
