package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/domain"
)

func TestDurationJsonMarshal(t *testing.T) {
	for _, d := range []time.Duration{
		0,
		3*time.Second + 500*time.Millisecond,
		-30 * time.Hour,
	} {
		jsonStr, err := json.Marshal(Duration{Duration: d})
		assert.NoError(t, err)
		var parsed Duration
		assert.NoError(t, json.Unmarshal(jsonStr, &parsed))
		assert.Equal(t, d, parsed.Duration)
	}
}

func TestDurationJsonUnmarshal(t *testing.T) {
	var d Duration

	assert.NoError(t, json.Unmarshal([]byte(`"2.5s"`), &d))
	assert.Equal(t, 2*time.Second+500*time.Millisecond, d.Duration)

	assert.NoError(t, json.Unmarshal([]byte(`3.5`), &d))
	assert.Equal(t, 3*time.Second+500*time.Millisecond, d.Duration)

	assert.Regexp(t, `unknown unit "xyz" in duration "123xyz"`, json.Unmarshal([]byte(`"123xyz"`), &d).Error())
	assert.Regexp(t, `invalid duration`, json.Unmarshal([]byte(`true`), &d).Error())
}
