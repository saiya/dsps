package domain_test

import (
	"testing"
	"time"

	. "github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestParseJwtExp(t *testing.T) {
	exp, err := ParseJwtExp("1300819380") // Example value from RFC 7519
	assert.NoError(t, err)
	assert.Equal(t, 2011, exp.Time().UTC().Year())
	assert.Equal(t, time.Month(3), exp.Time().UTC().Month())
	assert.Equal(t, 22, exp.Time().UTC().Day())

	_, err = ParseJwtExp("invalid")
	assert.Equal(t, `Invalid exp claim: invalid (strconv.ParseInt: parsing "invalid": invalid syntax)`, err.Error())
}

func TestJwtExpStringer(t *testing.T) {
	exp, err := ParseJwtExp("1300819380") // Example value from RFC 7519
	assert.NoError(t, err)
	assert.Equal(t, "1300819380", exp.String())
}

func TestNoneAlg(t *testing.T) {
	assert.True(t, JwtAlg("none").IsNone())
	assert.False(t, JwtAlg("RS256").IsNone())
}
