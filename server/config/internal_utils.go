package config

import (
	"time"

	"golang.org/x/xerrors"

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
		panic(xerrors.Errorf("makeDuration(\"%s\") resulted in %w", str, err))
	}
	return domain.Duration{Duration: d}
}

func makeDurationPtr(str string) *domain.Duration {
	d := makeDuration(str)
	return &d
}

func makeRegex(str string) *domain.Regex {
	result, err := domain.NewRegex(str)
	if err != nil {
		panic(xerrors.Errorf("makeRegex(\"%s\") resulted in %w", str, err))
	}
	return result
}

func intMustBeLargerThanZero(name string, value int) error {
	if value <= 0 {
		return xerrors.Errorf("%s must not be negative", name)
	}
	return nil
}

func durationMustBeLargerThanZero(name string, d domain.Duration) error {
	if d.Duration <= 0 {
		return xerrors.Errorf("%s must not be negative", name)
	}
	return nil
}
