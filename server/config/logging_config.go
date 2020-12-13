package config

// LoggingConfig represents logger and logging settings
type LoggingConfig struct {
	Debug      bool              `json:"debug"`
	Category   map[string]string `json:"category"`
	Attributes map[string]string `json:"attributes"`
}

var loggingConfigDefault = LoggingConfig{
	Debug:      false,
	Category:   map[string]string{},
	Attributes: map[string]string{},
}
