package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/saiya/dsps/server/domain"
)

func makeIntPtr(value int) *int {
	return &value
}

func makeFloat64Ptr(value float64) *float64 {
	return &value
}

func makeDuration(str string) domain.Duration {
	d, err := time.ParseDuration(str)
	if err != nil {
		panic(fmt.Sprintf("makeDuration(\"%s\") resulted in %v", str, err))
	}
	return domain.Duration{Duration: d}
}

func makeDurationPtr(str string) *domain.Duration {
	d := makeDuration(str)
	return &d
}

func intMustBeLargerThanZero(name string, value int) error {
	if value <= 0 {
		return errors.New("%s must not be negative")
	}
	return nil
}

func durationMustBeLargerThanZero(name string, d domain.Duration) error {
	if d.Duration <= 0 {
		return errors.New("%s must not be negative")
	}
	return nil
}
