package config

// LoggingConfig represents logger and logging settings
type LoggingConfig struct {
	Debug            bool              `json:"debug"`
	AuthRejectionLog bool              `json:"authRejectionLog"`
	Attributes       map[string]string `json:"attributes"`
}

var loggingConfigDefault = LoggingConfig{
	Debug:            false,
	AuthRejectionLog: true,
	Attributes:       map[string]string{},
}
