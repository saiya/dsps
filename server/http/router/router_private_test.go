package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcatPath(t *testing.T) {
	for _, testdata := range []struct {
		expected string
		data     []string
	}{
		{expected: "/", data: []string{}},
		{expected: "/", data: []string{""}},
		{expected: "/", data: []string{"/"}},
		{expected: "/", data: []string{"/", "/"}},
		{expected: "/", data: []string{"", "/", "", "/", ""}},
		{expected: "/a", data: []string{"a"}},
		{expected: "/a", data: []string{"a/"}},
		{expected: "/a", data: []string{"/a"}},
		{expected: "/a", data: []string{"/a/"}},
		{expected: "/a", data: []string{"/", "a"}},
		{expected: "/a", data: []string{"/", "a/"}},
		{expected: "/a", data: []string{"/", "/a"}},
		{expected: "/a", data: []string{"/", "/a/"}},
		{expected: "/a", data: []string{"/", "/a/", ""}},
		{expected: "/a/b", data: []string{"a", "b"}},
		{expected: "/a/b/c", data: []string{"a", "/b", "/c"}},
	} {
		assert.Equal(t, testdata.expected, concatPath(testdata.data...))
	}
}
