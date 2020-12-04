package domain

import (
	"encoding/json"
	"time"

	"golang.org/x/xerrors"
)

// Time wrapper struct
type Time struct {
	time.Time
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
