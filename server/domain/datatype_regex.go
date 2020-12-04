package domain

import (
	"encoding/json"
	"regexp"

	"golang.org/x/xerrors"
)

// Regex wrapper struct
type Regex struct {
	raw        string
	compiled   *regexp.Regexp
	groupNames []string
}

// NewRegex compiles regex
func NewRegex(regex string) (*Regex, error) {
	result := &Regex{}
	err := result.init(regex)
	return result, err
}

func (regex *Regex) init(str string) error {
	compiled, err := regexp.Compile(str)
	if err != nil {
		return xerrors.Errorf("Unable to parse Regex /%s/ : %w", str, err)
	}

	compiled.Longest()
	regex.raw = str
	regex.compiled = compiled
	regex.groupNames = compiled.SubexpNames()[1:]
	return nil
}

// String represents original regex
func (regex *Regex) String() string {
	return regex.raw
}

// GroupNames returns list of group (subregex) names
func (regex *Regex) GroupNames() []string {
	return regex.groupNames
}

// Match against given string and returns map of named groups. If not match returns nil.
func (regex *Regex) Match(entire bool, str string) map[string]string {
	compiled := regex.compiled
	captured := compiled.FindStringSubmatch(str)
	if captured == nil || (entire && len(captured[0]) != len(str)) {
		return nil
	}

	result := make(map[string]string, compiled.NumSubexp())
	for i, name := range compiled.SubexpNames()[1:] {
		result[name] = captured[i+1]
	}
	return result
}

// MarshalJSON method for configuration marshal/unmarshal
func (regex Regex) MarshalJSON() ([]byte, error) {
	return json.Marshal(regex.String())
}

// UnmarshalJSON method for configuration marshal/unmarshal
func (regex *Regex) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		return regex.init(value)
	default:
		return xerrors.New("invalid regex")
	}
}
