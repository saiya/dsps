package outgoing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/saiya/dsps/server/sentry"
	"github.com/stretchr/testify/assert"
)

func TestRetryResponseHandling(t *testing.T) {
	responses := []*mockResponse{
		newMockResponse(200, []byte("error while reading body")),
		newMockResponse(201, []byte("error while close")),
		newMockResponse(202, []byte("test response body")),
	}
	responses[0].body.ErrOnRead = errors.New("error while reading")
	responses[1].body.ErrOnClose = errors.New("error while closing")
	for _, res := range responses {
		attempts := 0
		assert.NoError(t, (&retry{}).Do(context.Background(), sentry.NewEmptySentry(), "test", func() (*http.Request, *http.Response, error) {
			attempts++
			return nil, &res.Response, nil
		}))
		assert.Equal(t, 1, attempts)
		res.assertProperlyClosed(t)
	}
}

func TestRetrySuccessForUnexpectedHttpStatus(t *testing.T) {
	resQueue := []*mockResponse{
		newMockResponse(500, []byte("Internal server error")),
		newMockResponse(404, []byte("Not found")), // 1st retry
		newMockResponse(201, []byte{}),            // 2nd retry
	}

	attempts := 0
	assert.NoError(t, (&retry{
		count: 2,
	}).Do(context.Background(), sentry.NewEmptySentry(), "test", func() (*http.Request, *http.Response, error) {
		attempts++
		return nil, &resQueue[attempts-1].Response, nil
	}))
	assert.Equal(t, 3, attempts)
}

func TestRetryFailureForUnexpectedHttpStatus(t *testing.T) {
	resQueue := []*mockResponse{
		newMockResponse(500, []byte("Internal server error")),
		newMockResponse(404, []byte("Not found")), // 1st retry
		newMockResponse(404, []byte("Not found")), // 2nd retry
		newMockResponse(201, []byte{}),            // 3rd retry
	}

	attempts := 0
	err := (&retry{
		count: 2, // Give up after 2nd retry
	}).Do(context.Background(), sentry.NewEmptySentry(), "test", func() (*http.Request, *http.Response, error) {
		attempts++
		return nil, &resQueue[attempts-1].Response, nil
	})
	assert.Error(t, err)
	assert.Regexp(t, `status code 404 returned`, err.Error())
	assert.Equal(t, 3, attempts)
}

func TestRetrySuccessForError(t *testing.T) {
	attempts := 0
	assert.NoError(t, (&retry{
		count: 2,
	}).Do(context.Background(), sentry.NewEmptySentry(), "test", func() (*http.Request, *http.Response, error) {
		attempts++
		if attempts <= 2 {
			return nil, nil, errors.New("test error")
		}
		return nil, &newMockResponse(200, []byte{}).Response, nil
	}))
	assert.Equal(t, 3, attempts)
}

func TestRetryFailureForError(t *testing.T) {
	testError := errors.New("test error")
	attempts := 0
	err := (&retry{
		count: 2,
	}).Do(context.Background(), sentry.NewStubSentry(), "test", func() (*http.Request, *http.Response, error) {
		attempts++
		return &http.Request{}, nil, testError
	})
	assert.Equal(t, testError, err)
	assert.Equal(t, 3, attempts)
}

func TestRetryWait(t *testing.T) {
	retryWithoutJitter := &retry{
		interval:           3 * time.Second,
		intervalMultiplier: 1.5,
		intervalJitter:     0,
	}
	assert.InDelta(t, 3.0, retryWithoutJitter.computeRetryWait(0).Seconds(), 0.01)
	assert.InDelta(t, 3.0*1.5, retryWithoutJitter.computeRetryWait(1).Seconds(), 0.01)
	assert.InDelta(t, 3.0*1.5*1.5, retryWithoutJitter.computeRetryWait(2).Seconds(), 0.01)
	assert.InDelta(t, 3.0*1.5*1.5*1.5, retryWithoutJitter.computeRetryWait(3).Seconds(), 0.01)
	assert.Equal(t, retryWithoutJitter.computeRetryWait(0).Seconds(), retryWithoutJitter.computeRetryWait(0).Seconds())

	retryWithJitter := &retry{
		interval:           3 * time.Second,
		intervalMultiplier: 1.5,
		intervalJitter:     100 * time.Millisecond,
	}
	assert.InDelta(t, 3.0*1.5*1.5*1.5, retryWithJitter.computeRetryWait(3).Seconds(), 0.11)
	assert.NotEqual(t, retryWithJitter.computeRetryWait(3).Seconds(), retryWithJitter.computeRetryWait(3).Seconds())
}

func TestPostprocess(t *testing.T) {
	r := &retry{}

	someError := errors.New("some error")
	shouldRetry, err := r.postprocess(nil, someError)
	assert.True(t, shouldRetry)
	assert.NotNil(t, err, "postprocess() must return non-nil err")
	assert.Equal(t, someError, err)

	shouldRetry, err = r.postprocess(nil, nil)
	assert.True(t, shouldRetry)
	assert.NotNil(t, err, "postprocess() must return non-nil err")
	assert.Equal(t, "no response object returned", err.Error())

	for statusCode, expectedRetry := range map[int]bool{
		// Explicitly listed as shouldRetry == true
		400: true,
		// Explicitly listed as shouldRetry == false
		401: false,
		// Default
		499: true,
	} {
		res := &http.Response{StatusCode: statusCode}
		shouldRetry, err = r.postprocess(res, nil)
		assert.Equal(t, expectedRetry, shouldRetry)
		assert.NotNil(t, err, "postprocess() must return non-nil err")
		assert.Equal(t, fmt.Sprintf("status code %d returned", statusCode), err.Error())
	}
}
