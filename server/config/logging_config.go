package config

// LoggingConfig represents logger and logging settings
type LoggingConfig struct {
	Debug      bool              `json:"debug"`
	Attributes map[string]string `json:"attributes"`
}

var loggingConfigDefault = LoggingConfig{
	Debug:      false,
	Attributes: map[string]string{},
}
