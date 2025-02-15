package pongo2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAttributes_String(t *testing.T) {
	testCases := []struct {
		input *Attributes
		want  string
	}{
		{input: newAttributes([][2]string{}), want: ""},
		{input: newAttributes([][2]string{
			{"key1", "value1"},
			{"key2", "value2"},
			{"data-aaa", "value3"},
		}), want: `key1="value1" key2="value2" data-aaa="value3"`},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.want, tc.input.String())
	}
}
