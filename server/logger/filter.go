package logger

import (
	"sync"
)

// Filter controls log verbosity
type Filter struct {
	// Do not modify after creation otherwise breaks goroutine-safe
	thresholds       sync.Map
	defaultThreshold Level
}

func newDefaultFilter() *Filter {
	return &Filter{
		thresholds:       sync.Map{},
		defaultThreshold: INFO,
	}
}

// NewFilter creates Filter instance with given thresholds configuration.
func NewFilter(thresholds map[string]string) (*Filter, error) {
	result := Filter{
		thresholds:       sync.Map{},
		defaultThreshold: INFO,
	}
	for key, value := range thresholds {
		level, err := ParseLevel(value)
		if err != nil {
			return nil, err
		}
		result.SetThreshold(ParseCategory(key), level)
	}
	return &result, nil
}

// Filter determines whether to output (true) the log or not (false).
func (filter *Filter) Filter(level Level, cat Category) bool {
	threshold, ok := filter.thresholds.Load(cat)
	if !ok {
		threshold = filter.defaultThreshold
	}
	return level >= threshold.(Level)
}

// SetThreshold changes threshold immediately
func (filter *Filter) SetThreshold(cat Category, level Level) {
	if cat == "*" {
		filter.defaultThreshold = level
	} else {
		filter.thresholds.Store(cat, level)
	}
}
