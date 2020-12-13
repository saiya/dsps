package domain

import (
	"fmt"
	"strconv"
	"time"
)

// JwtExp is "exp" claim, number of seconds from 1970-01-01T00:00:00Z UTC without leap seconds.
type JwtExp time.Time // Intentionally use time.Time rather than domain.Time to prevent using JSON marshaler of domain.Time

// ParseJwtExp parses claim value
func ParseJwtExp(str string) (JwtExp, error) {
	epoch, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return JwtExp{}, fmt.Errorf("Invalid exp claim: %s (%w)", str, err)
	}
	return JwtExp(time.Unix(epoch, 0)), nil
}

// String method to implement Stringer
func (exp JwtExp) String() string {
	return strconv.FormatInt(time.Time(exp).Unix(), 10)
}

// Time returns time.Time instance equivalent to this
func (exp JwtExp) Time() time.Time {
	return time.Time(exp)
}

// JwtAlg is "alg" claim, signing algorithm of a JWT
type JwtAlg string

// IsNone returns true only for "none" alg.
func (alg JwtAlg) IsNone() bool {
	return alg == "none"
}

// JwtJti is "jti" claim, ID of a JWT
type JwtJti string

// JwtIss is "iss" claim, issuer of a JWT
type JwtIss string

// JwtAud is "aud" claim, recipient of a JWT
type JwtAud string
