package domain

import (
	"encoding/json"
	"strings"
	"text/template"

	"golang.org/x/xerrors"
)

// TemplateString is Template wrappwer struct
type TemplateString struct {
	template       *template.Template
	templateString string
}

// NewTemplateString initialize template
func NewTemplateString(value string) (*TemplateString, error) {
	result := &TemplateString{}
	return result, result.init(value)
}

func (tpl *TemplateString) init(value string) error {
	parsed, err := template.New("template-string").Option("missingkey=error").Parse(value)
	if err != nil {
		return xerrors.Errorf("Unable to parse Template \"%s\" %w", value, err)
	}
	tpl.template = parsed
	tpl.templateString = value
	return nil
}

// String returns original template string
func (tpl TemplateString) String() string {
	return tpl.templateString
}

// Execute evaluates template
func (tpl TemplateString) Execute(data interface{}) (string, error) {
	output := strings.Builder{}
	if err := tpl.template.Execute(&output, data); err != nil {
		return "", err
	}
	return output.String(), nil
}

// MarshalJSON method for configuration marshal/unmarshal
func (tpl TemplateString) MarshalJSON() ([]byte, error) {
	return json.Marshal(tpl.templateString)
}

// UnmarshalJSON method for configuration marshal/unmarshal
func (tpl *TemplateString) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		return tpl.init(value)
	default:
		return xerrors.New("invalid template")
	}
}
