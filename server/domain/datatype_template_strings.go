package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

// TemplateStrings is list of TemplateString
type TemplateStrings struct {
	Templates []TemplateString
}

// NewTemplateStrings initialize templates
func NewTemplateStrings(templates ...TemplateString) TemplateStrings {
	return TemplateStrings{Templates: templates}
}

// String returns original template strings
func (ts TemplateStrings) String() string {
	list := make([]string, len(ts.Templates))
	for i, tpl := range ts.Templates {
		list[i] = fmt.Sprintf(`"%s"`, tpl.String())
	}
	return fmt.Sprintf("TemplateStrings{%s}", strings.Join(list, ", "))
}

// Execute evaluates templates
func (ts TemplateStrings) Execute(data TemplateStringEnv) ([]string, error) {
	var err error
	result := make([]string, len(ts.Templates))
	for i, tpl := range ts.Templates {
		result[i], err = tpl.Execute(data)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate template #%d: %w", i+1, err)
		}
	}
	return result, err
}

// MarshalJSON method for configuration marshal/unmarshal
func (ts TemplateStrings) MarshalJSON() ([]byte, error) {
	list := make([]string, len(ts.Templates))
	for i, tpl := range ts.Templates {
		list[i] = tpl.templateString
	}
	return json.Marshal(list)
}

// UnmarshalJSON method for configuration marshal/unmarshal
func (ts *TemplateStrings) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		tpl, err := NewTemplateString(value)
		if err != nil {
			return err
		}
		*ts = NewTemplateStrings(tpl)
		return nil
	case []interface{}:
		tpls := make([]TemplateString, len(value))
		for i, element := range value {
			switch item := element.(type) {
			case string:
				tpl, err := NewTemplateString(item)
				if err != nil {
					return fmt.Errorf("failed to parse template #%d: %w", i+1, err)
				}
				tpls[i] = tpl
			default:
				return xerrors.Errorf("invalid templates, expected string or list of string")
			}
		}
		*ts = NewTemplateStrings(tpls...)
		return nil
	default:
		return xerrors.New("invalid templates, expected string or list of string")
	}
}
