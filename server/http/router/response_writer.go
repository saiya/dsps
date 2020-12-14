package router

import "net/http"

// ResponseWriter extends net/http ResponseWriter
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)

	Written() ResponseWritten
}

// ResponseWritten is summary of written data
type ResponseWritten struct {
	StatusCode int
	BodyBytes  int
}

type responseWriter struct {
	inner http.ResponseWriter

	written ResponseWritten
}

// NewResponseWriter wraps writer
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		inner: w,
		written: ResponseWritten{
			StatusCode: 200, // default value of http.ResponseWriter
			BodyBytes:  0,
		},
	}
}

func (w *responseWriter) Header() http.Header {
	return w.inner.Header()
}

func (w *responseWriter) Write(body []byte) (int, error) {
	bytes, err := w.inner.Write(body)
	w.written.BodyBytes += bytes
	return bytes, err
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.inner.WriteHeader(statusCode)
	w.written.StatusCode = statusCode
}

func (w *responseWriter) Written() ResponseWritten {
	return w.written
}
