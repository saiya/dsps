package testing

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// WithTextFile creates temporary text file
func WithTextFile(t *testing.T, content string, f func(filename string)) {
	file, err := ioutil.TempFile(os.TempDir(), "dsps-test-*.txt")
	assert.NoError(t, err)
	defer assert.NoError(t, os.Remove(file.Name()))

	assert.NoError(t, ioutil.WriteFile(file.Name(), []byte(content), os.ModePerm))
	f(file.Name())
}
