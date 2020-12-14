package config

// LoggingConfig represents logger and logging settings
type LoggingConfig struct {
	Category   map[string]string `json:"category"`
	Attributes map[string]string `json:"attributes"`
}

func loggingConfigDefault() *LoggingConfig {
	return &LoggingConfig{
		Category:   map[string]string{},
		Attributes: map[string]string{},
	}
}
