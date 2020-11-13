package testing

import (
	"fmt"
	"time"

	"github.com/saiya/dsps/server/domain"
)

// MakeIntPtr returns *int
func MakeIntPtr(value int) *int {
	return &value
}

// MakeDuration parses time.Duration
func MakeDuration(str string) domain.Duration {
	d, err := time.ParseDuration(str)
	if err != nil {
		panic(fmt.Sprintf("makeDuration(\"%s\") resulted in %v", str, err))
	}
	return domain.Duration{Duration: d}
}

// MakeDurationPtr parses *time.Duration
func MakeDurationPtr(str string) *domain.Duration {
	d := MakeDuration(str)
	return &d
}
