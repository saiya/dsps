package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/domain"
)

func TestSingleTemplateStringsJSONMapping(t *testing.T) {
	jsonStr := `{"tpls":"chat-room-{{.id}}"}`
	var parsed struct {
		Tpls TemplateStrings `json:"tpls"`
	}
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))

	result, err := parsed.Tpls.Execute(map[string]string{"id": "1234"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"chat-room-1234"}, result)

	generated, err := json.Marshal(parsed)
	assert.NoError(t, err)
	assert.Equal(t, `{"tpls":["chat-room-{{.id}}"]}`, string(generated))
}

func TestMultipleTemplateStringsJSONMapping(t *testing.T) {
	jsonStr := `{"tpls":["chat-room-{{.id}}","another-room-{{.id}}"]}`
	var parsed struct {
		Tpls TemplateStrings `json:"tpls"`
	}
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))

	result, err := parsed.Tpls.Execute(map[string]string{"id": "1234"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"chat-room-1234", "another-room-1234"}, result)

	generated, err := json.Marshal(parsed)
	assert.NoError(t, err)
	assert.Equal(t, jsonStr, string(generated))
}

func TestInvalidTemplateStringsJSON(t *testing.T) {
	var tpls TemplateStrings
	assert.Regexp(t, `Unable to parse Template`, json.Unmarshal([]byte(`"chat-room-{{.channel.id}"`), &tpls).Error())
	assert.Regexp(t, `failed to parse template #1: Unable to parse Template`, json.Unmarshal([]byte(`["chat-room-{{.channel.id}"]`), &tpls).Error())
	assert.Regexp(t, `invalid templates, expected string or list of string`, json.Unmarshal([]byte(`["foo", 1234]`), &tpls).Error())
	assert.Regexp(t, `invalid templates, expected string or list of string`, json.Unmarshal([]byte(`1234`), &tpls).Error())
}
