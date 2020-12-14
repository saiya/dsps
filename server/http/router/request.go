package router

import (
	"io/ioutil"
	"net/http"
)

// Request wraps http.Request
type Request struct {
	*http.Request
}

// ReadBody read request body
func (req Request) ReadBody() ([]byte, error) {
	// > The Server will close the request body. The ServeHTTP Handler does not need to.
	// https://golang.org/pkg/net/http/#ResponseWriter
	return ioutil.ReadAll(req.Body)
}

// GetQueryParam returns URL query parameter or ""
func (req Request) GetQueryParam(name string) string {
	return req.GetQueryParamOrDefault(name, "")
}

// GetQueryParamOrDefault returns URL query parameter or ""
func (req Request) GetQueryParamOrDefault(name string, defaultValue string) string {
	list := req.URL.Query()[name]
	if len(list) > 0 {
		return list[0]
	}
	return defaultValue
}
