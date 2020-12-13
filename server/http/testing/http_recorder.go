package testing

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
)

// AssertRecordedCode assert DSPS standard HTTP error response
func AssertRecordedCode(t *testing.T, rec *httptest.ResponseRecorder, httpStatusCode int, dspsCode domain.ErrorWithCode) {
	assert.Equal(t, httpStatusCode, rec.Code)
	assert.Equal(t, dspsCode.Code(), GetBodyJSONMap(t, rec)["code"], `expected DSPS code %v but response BODY is %s`, dspsCode, string(rec.Body.Bytes()))
}

// GetBodyJSONMap extract JSON from response body
func GetBodyJSONMap(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	body := make(map[string]interface{})
	GetBodyJSON(t, rec, &body)
	return body
}

// GetBodyJSON extract JSON from response body
func GetBodyJSON(t *testing.T, rec *httptest.ResponseRecorder, body interface{}) {
	assert.Regexp(t, "application/json(;|$)", rec.Header().Get("Content-Type"))
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), body))
}
