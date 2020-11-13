package config

import (
	"fmt"
	"time"

	"github.com/saiya/dsps/server/domain"
)

func makeIntPtr(value int) *int {
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
