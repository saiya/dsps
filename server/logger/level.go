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

func (level Level) String() string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return fmt.Sprintf("logger.Level(%d)", level)
}

// ParseLevel parses log level string
func ParseLevel(str string) (Level, error) {
	given := strings.ToUpper(str)
	for i := 0; i <= int(FATAL); i++ {
		if given == Level(i).String() {
			return Level(i), nil
		}
	}
	return DEBUG, fmt.Errorf(`invalid log level string given: "%s"`, str)
}
