package logger

// Category of logs
type Category string

const (
	// CatServer is server lifecycle log
	CatServer Category = "server"
	// CatLogger is logging system events
	CatLogger = "logger"
	// CatTracing is tracing system events
	CatTracing = "tracing"
	// CatAuth is auth related events
	CatAuth = "auth"
	// CatHTTP is HTTP layer log
	CatHTTP = "http"
	// CatStorage is storage events
	CatStorage = "storage"
	// CatOutgoingWebhook is outgoing webhook events
	CatOutgoingWebhook = "webhook-out"
)

// ParseCategory make category string
func ParseCategory(str string) Category {
	return Category(str)
}
