package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/domain"
)

func TestCIDRStrings(t *testing.T) {
	str := "192.168.0.0/16"
	cidr, err := ParseCIDR(str)
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.0/16", cidr.String())
	assert.NotNil(t, cidr.IPNet())

	_, err = ParseCIDR("192.168.0.0/33")
	assert.Contains(t, err.Error(), `invalid CIDR address: 192.168.0.0/33`)
}

func TestCIDRJSONMarshal(t *testing.T) {
	str := "192.168.0.0/16"
	cidr, err := ParseCIDR(str)
	assert.NoError(t, err)
	jsonBytes, err := json.Marshal(cidr)
	assert.NoError(t, err)
	assert.Equal(t, `"192.168.0.0/16"`, string(jsonBytes))
}

func TestCIDRJSONUnMarshal(t *testing.T) {
	var cidr CIDR
	assert.NoError(t, json.Unmarshal([]byte(`"192.168.0.0/16"`), &cidr))
	assert.Equal(t, "192.168.0.0/16", cidr.String())
	assert.True(t, cidr.Contains("192.168.0.1"))

	assert.Error(t, json.Unmarshal([]byte(`INVALID-JSON`), &cidr))
	assert.Contains(t, json.Unmarshal([]byte(`"192.168.0.0/33"`), &cidr).Error(), `invalid CIDR address: 192.168.0.0/33`)
	assert.Contains(t, json.Unmarshal([]byte(`true`), &cidr).Error(), `invalid CIDR notation`)
}

func TestCIDRCalc(t *testing.T) {
	str := "192.168.0.0/24"
	cidr, err := ParseCIDR(str)
	assert.NoError(t, err)

	assert.True(t, cidr.Contains("192.168.0.1"))
	assert.False(t, cidr.Contains("192.168.1.1"))
	assert.False(t, cidr.Contains("2001:db8::/32"))
	assert.False(t, cidr.Contains("INVALID-IP"))
}
