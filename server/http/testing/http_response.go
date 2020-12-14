package testing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

// AssertErrorResponse ensure DSPS standard error response
func AssertErrorResponse(t *testing.T, res *http.Response, httpStatus int, dspsError domain.ErrorWithCode, messageRegex string) {
	assert.Equal(t, httpStatus, res.StatusCode)

	body := BodyJSONMapOfRes(t, res)
	assert.Regexp(t, messageRegex, body["error"])

	var expectedCode interface{} = nil
	if dspsError != nil {
		expectedCode = dspsError.Code()
	}
	assert.Equal(t, expectedCode, body["code"], `expected code: %s but response body is: %v`, expectedCode, body)
}

// AssertInternalServerErrorResponse ensure 500 response
func AssertInternalServerErrorResponse(t *testing.T, res *http.Response) {
	assert.Equal(t, 500, res.StatusCode)
	assert.NoError(t, res.Body.Close())
}

// AssertResponseJSON ensure response body JSON content
func AssertResponseJSON(t *testing.T, res *http.Response, httpStatus int, expected map[string]interface{}) map[string]interface{} {
	assert.Equal(t, httpStatus, res.StatusCode)

	body := BodyJSONMapOfRes(t, res)
	for key, value := range expected {
		assert.Equal(t, value, body[key])
	}
	return body
}

// BodyJSONMapOfRes extract JSON from response body
func BodyJSONMapOfRes(t *testing.T, res *http.Response) map[string]interface{} {
	body := make(map[string]interface{})
	BodyJSONOfRes(t, res, &body)
	return body
}

// BodyJSONOfRes extract JSON from response body
func BodyJSONOfRes(t *testing.T, res *http.Response, body interface{}) {
	assert.Regexp(t, "application/json(;|$)", res.Header.Get("Content-Type"))

	raw, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.NoError(t, res.Body.Close())
	assert.NoError(t, json.Unmarshal(raw, body), "failed to parse response JSON: %s", string(raw))
}
