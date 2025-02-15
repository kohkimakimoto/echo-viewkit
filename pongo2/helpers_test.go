package pongo2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertToCamelCase(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "key", want: "key"},
		{input: "key-value", want: "keyValue"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.want, convertToCamelCase(tc.input))
	}
}

func TestConvertToKebabCase(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "key", want: "key"},
		{input: "keyValue", want: "key-value"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.want, convertToKebabCase(tc.input))
	}
}
