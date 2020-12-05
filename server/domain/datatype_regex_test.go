package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/domain"
)

func TestNewRegex(t *testing.T) {
	regex, err := NewRegex("chat-room-(?P<id>\\d+)")
	assert.NoError(t, err)

	assert.Equal(t, "chat-room-(?P<id>\\d+)", regex.String())
	assert.Equal(t, []string{"id"}, regex.GroupNames())

	_, err = NewRegex("chat-room-(?P<id>\\d+")
	assert.EqualError(t, err, "Unable to parse Regex /chat-room-(?P<id>\\d+/ : error parsing regexp: missing closing ): `chat-room-(?P<id>\\d+`")
}

func TestRegexWithoutGroup(t *testing.T) {
	regex, err := NewRegex("chat-room-\\d+")
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{}, regex.Match(true, "chat-room-123"))
	assert.NotNil(t, regex.Match(true, "chat-room-123")) // Should be distingushable with unmatch
	assert.Equal(t, map[string]string{}, regex.Match(false, "chat-room-123"))
	assert.Nil(t, regex.Match(true, "chat-room-XYZ"))                             // No match
	assert.Nil(t, regex.Match(false, "chat-room-XYZ"))                            // No match
	assert.Nil(t, regex.Match(true, "**chat-room-123**"))                         // No entire match
	assert.Equal(t, map[string]string{}, regex.Match(false, "**chat-room-123**")) // Partial match
}

func TestRegexMatch(t *testing.T) {
	regex, err := NewRegex("chat-room-(?P<id>\\d+)")
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{"id": "123"}, regex.Match(true, "chat-room-123"))
	assert.Equal(t, map[string]string{"id": "123"}, regex.Match(false, "chat-room-123"))
	assert.Nil(t, regex.Match(true, "chat-room-XYZ"))                                        // No match
	assert.Nil(t, regex.Match(false, "chat-room-XYZ"))                                       // No match
	assert.Nil(t, regex.Match(true, "**chat-room-123**"))                                    // No entire match
	assert.Equal(t, map[string]string{"id": "123"}, regex.Match(false, "**chat-room-123**")) // Partial match
}

func TestRegexJsonMapping(t *testing.T) {
	jsonStr := `{"regex":"chat-room-(?P\u003cid\u003e\\d+)"}`
	var parsed struct {
		Regex Regex `json:"regex"`
	}
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))
	assert.Equal(t, map[string]string{"id": "123"}, parsed.Regex.Match(true, "chat-room-123"))

	generated, err := json.Marshal(parsed)
	assert.NoError(t, err)
	assert.Equal(t, jsonStr, string(generated))
}
