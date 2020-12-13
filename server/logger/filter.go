package logger

// Filter controls log verbosity
type Filter struct {
	// Do not modify after creation otherwise breaks goroutine-safe
	thresholds       map[Category]Level
	defaultThreshold Level
}

func newDefaultFilter() *Filter {
	return &Filter{
		thresholds:       map[Category]Level{},
		defaultThreshold: INFO,
	}
}

// NewFilter creates Filter instance with given thresholds configuration.
func NewFilter(thresholds map[string]string) (*Filter, error) {
	result := Filter{
		thresholds:       map[Category]Level{},
		defaultThreshold: INFO,
	}
	for key, value := range thresholds {
		level, err := ParseLevel(value)
		if err != nil {
			return nil, err
		}
		if key == "*" {
			result.defaultThreshold = level
		} else {
			result.thresholds[ParseCategory(key)] = level
		}
	}
	return &result, nil
}

// Filter determines whether to output (true) the log or not (false).
func (filter *Filter) Filter(level Level, cat Category) bool {
	threshold, ok := filter.thresholds[cat]
	if !ok {
		threshold = filter.defaultThreshold
	}
	return level >= threshold
}
