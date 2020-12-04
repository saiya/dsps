package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/domain"
)

func TestNewTemplateString(t *testing.T) {
	tpl, err := NewTemplateString("chat-room-{{.regex.id}}")
	assert.NoError(t, err)

	result, err := tpl.Execute(map[string](map[string]string){
		"regex": map[string]string{"id": "1234"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "chat-room-1234", result)
}

func TestTemplateStringJsonMapping(t *testing.T) {
	jsonStr := `{"tpl":"chat-room-{{.id}}"}`
	var parsed struct {
		Tpl TemplateString `json:"tpl"`
	}
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))

	result, err := parsed.Tpl.Execute(map[string]string{"id": "1234"})
	assert.NoError(t, err)
	assert.Equal(t, "chat-room-1234", result)

	generated, err := json.Marshal(parsed)
	assert.NoError(t, err)
	assert.Equal(t, jsonStr, string(generated))
}
