package logger

import (
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

// LogCapture represents log capture system for testing purpose
type LogCapture struct {
	lock       sync.Mutex
	logEntries []CapturedLog
}

func (c *LogCapture) append(logEntry CapturedLog) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.logEntries = append(c.logEntries, logEntry)
}

// LastLog returns captured log
func (c *LogCapture) LastLog(offset int) *CapturedLog {
	c.lock.Lock()
	defer c.lock.Unlock()
	return &c.logEntries[len(c.logEntries)-1-offset]
}

// Count returns total count of captured logs
func (c *LogCapture) Count() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return len(c.logEntries)
}

// CapturedLog represents captured log object
type CapturedLog struct {
	zapcore.Entry
	Fields []zapcore.Field
}

// FindErrorField returns error field of the log or nil
func (log CapturedLog) FindErrorField() error {
	for _, f := range log.Fields {
		if err, ok := f.Interface.(error); ok && f.Type == zapcore.ErrorType {
			return err
		}
	}
	return nil
}

// FindStringField string field of the given name
func (log CapturedLog) FindStringField(key string) string {
	for _, f := range log.Fields {
		if f.Key == key && f.Type == zapcore.StringType {
			return f.String
		}
	}
	return ""
}

// WithTestLogger swaps global logger for testing only.
func WithTestLogger(t *testing.T, filter *Filter, f func(lc *LogCapture)) {
	prev := rootLogger
	defer func() { rootLogger = prev }()

	if filter == nil {
		filter = rootLogger.filter
	}

	lc := LogCapture{}
	// based on uber-go/fx Sentry example: https://github.com/uber-go/zap/issues/418#issuecomment-438323524
	logger := zaptest.NewLogger(t).WithOptions(zap.WrapCore(func(z zapcore.Core) zapcore.Core {
		return zapcore.NewTee(z, lc.createZapCore())
	}))
	rootLogger = &loggerImpl{
		zap:    logger,
		filter: filter,
	}
	f(&lc)
}

type logCaptureZapCore struct {
	*LogCapture
	fields []zapcore.Field
}

func (c *LogCapture) createZapCore() zapcore.Core {
	return &logCaptureZapCore{LogCapture: c}
}

func (c *logCaptureZapCore) Sync() error {
	return nil
}

func (c *logCaptureZapCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *logCaptureZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	c.append(CapturedLog{
		Entry:  entry,
		Fields: append(c.fields, fields...),
	})
	return nil
}

func (c *logCaptureZapCore) With(fields []zapcore.Field) zapcore.Core {
	return &logCaptureZapCore{
		LogCapture: c.LogCapture,
		fields:     append(c.fields, fields...),
	}
}

func (c *logCaptureZapCore) Enabled(level zapcore.Level) bool {
	return true
}
