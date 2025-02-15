package pongo2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseSlotDataExpr(t *testing.T) {
	// Test cases
	testCases := []struct {
		input  string
		result *slotData
		isErr  bool
	}{
		{
			input: "abcd1234",
			result: &slotData{
				name: "abcd1234",
				keys: []*slotDataKey{},
			},
		},
		{
			input: "{key1, key2, key3}",
			result: &slotData{
				name: "",
				keys: []*slotDataKey{
					{name: "key1", alias: ""},
					{name: "key2", alias: ""},
					{name: "key3", alias: ""},
				},
			},
		},
		{
			input: "{key1: alias1, key2: alias2, key3: alias3}",
			result: &slotData{
				name: "",
				keys: []*slotDataKey{
					{name: "key1", alias: "alias1"},
					{name: "key2", alias: "alias2"},
					{name: "key3", alias: "alias3"},
				},
			},
		},
		{
			input: "{key1: alias1, key2, key3: alias3}",
			result: &slotData{
				name: "",
				keys: []*slotDataKey{
					{name: "key1", alias: "alias1"},
					{name: "key2", alias: ""},
					{name: "key3", alias: "alias3"},
				},
			},
		},
		{
			input: "{}",
			result: &slotData{
				name: "",
				keys: []*slotDataKey{},
			},
		},
		{
			input: "{key1}",
			result: &slotData{
				name: "",
				keys: []*slotDataKey{
					{name: "key1", alias: ""},
				},
			},
		},
		{
			input: "{invalid,,key}",
			isErr: true,
		},
	}

	for _, tc := range testCases {
		result, err := parseSlotDataExpr(tc.input)
		if tc.isErr {
			assert.Error(t, err)
			assert.Nil(t, result)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.result, result)
		}
	}
}
