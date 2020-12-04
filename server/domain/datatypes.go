package domain

import (
	"encoding/json"
	"regexp"
	"text/template"
	"time"

	"golang.org/x/xerrors"
)

// Duration wrapper struct
type Duration struct {
	time.Duration
}

// Time wrapper struct
type Time struct {
	time.Time
}

// Regex wrapper struct
type Regex struct {
	*regexp.Regexp
}

// TemplateString is Template wrappwer struct
type TemplateString struct {
	*template.Template

	templateString string
}

// MarshalJSON method for configuration marshal/unmarshal
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON method for configuration marshal/unmarshal
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return xerrors.Errorf("Unable to parse Duration \"%s\" %w", value, err)
		}
		return nil
	default:
		return xerrors.New("invalid duration")
	}
}

// MarshalJSON method for configuration marshal/unmarshal
func (regex Regex) MarshalJSON() ([]byte, error) {
	return json.Marshal(regex.Regexp.String())
}

// UnmarshalJSON method for configuration marshal/unmarshal
func (regex *Regex) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		parsed, err := regexp.Compile(value)
		if err != nil {
			return xerrors.Errorf("Unable to parse Regex /%s/ %w", value, err)
		}
		regex.Regexp = parsed
		return nil
	default:
		return xerrors.New("invalid regex")
	}
}

func (tpl TemplateString) String() string {
	return tpl.templateString
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
		parsed, err := template.New("").Parse(value)
		if err != nil {
			return xerrors.Errorf("Unable to parse Template \"%s\" %w", value, err)
		}
		tpl.templateString = value
		tpl.Template = parsed
		return nil
	default:
		return xerrors.New("invalid template")
	}
}
