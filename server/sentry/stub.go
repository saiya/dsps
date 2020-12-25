package sentry

import (
	"sync"

	sentrygo "github.com/getsentry/sentry-go"
)

// StubSentry is stub (testing)
type StubSentry struct {
	lock        sync.Mutex
	errors      []error
	breadcrumbs []*sentrygo.Breadcrumb
	tags        map[string]string
	context     map[string]interface{}
}

// NewStubSentry creates stub implementation.
func NewStubSentry() *StubSentry {
	return &StubSentry{
		tags:    make(map[string]string),
		context: make(map[string]interface{}),
	}
}

// GetLastError returns last captured error
func (s *StubSentry) GetLastError() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.errors) == 0 {
		return nil
	}
	return s.errors[len(s.errors)-1]
}

// GetBreadcrumbs returns list of captured data
func (s *StubSentry) GetBreadcrumbs() []*sentrygo.Breadcrumb {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.breadcrumbs
}

// GetTags returns map of tags
func (s *StubSentry) GetTags() map[string]string {
	s.lock.Lock()
	defer s.lock.Unlock()
	result := make(map[string]string, len(s.tags))
	for key, value := range s.tags {
		result[key] = value
	}
	return result
}

// GetContext returns map of contexts
func (s *StubSentry) GetContext() map[string]interface{} {
	s.lock.Lock()
	defer s.lock.Unlock()
	result := make(map[string]interface{}, len(s.tags))
	for key, value := range s.context {
		result[key] = value
	}
	return result
}
