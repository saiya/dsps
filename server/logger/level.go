package logger

import (
	"fmt"
	"strings"
)

// Level represents log severity.
// Larger value represents more important log.
type Level int

const (
	// DEBUG log level
	DEBUG Level = iota
	// INFO log level
	INFO
	// WARN log level
	WARN
	// ERROR log level
	ERROR
	// FATAL log level
	FATAL
)

// ParseLevel parses log level string
func ParseLevel(str string) (Level, error) {
	switch strings.ToUpper(str) {
	case "DEBUG":
		return DEBUG, nil
	case "INFO":
		return INFO, nil
	case "WARN":
		return WARN, nil
	case "ERROR":
		return ERROR, nil
	case "FATAL":
		return FATAL, nil
	default:
		return DEBUG, fmt.Errorf(`invalid log level string given: "%s"`, str)
	}
}
