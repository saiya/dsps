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
	assert.Equal(t, dspsCode.Code(), BodyJSONMapOfRec(t, rec)["code"], `expected DSPS code %v but response BODY is %s`, dspsCode, string(rec.Body.Bytes()))
}

// BodyJSONMapOfRec extract JSON from response body
func BodyJSONMapOfRec(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	body := make(map[string]interface{})
	BodyJSONOfRec(t, rec, &body)
	return body
}

// BodyJSONOfRec extract JSON from response body
func BodyJSONOfRec(t *testing.T, rec *httptest.ResponseRecorder, body interface{}) {
	assert.Regexp(t, "application/json(;|$)", rec.Header().Get("Content-Type"))
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), body))
}
