package view

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseHTMLToTOC(t *testing.T) {
	testCases := []struct {
		html     string
		minDepth int
		maxDepth int
		expected []*TOCItem
		error    error
	}{
		{
			html: `
<h1 id="h1">Heading 1</h1>
<h2 id="h2">Heading 2</h2>
<h3 id="h3">Heading 3</h3>
`,
			minDepth: 0,
			maxDepth: 0,
			expected: []*TOCItem{
				{
					Level: 1,
					Id:    "h1",
					Title: "Heading 1",
					Children: []*TOCItem{
						{
							Level: 2,
							Id:    "h2",
							Title: "Heading 2",
							Children: []*TOCItem{
								{
									Level: 3,
									Id:    "h3",
									Title: "Heading 3",
								},
							},
						},
					},
				},
			},
			error: nil,
		},
		{
			html: `
<h1 id="h1">Heading 1</h1>
<h2 id="h2">Heading 2</h2>
<h3 id="h3">Heading 3</h3>
<h4 id="h4">Heading 4</h4>
`,
			minDepth: 2,
			maxDepth: 3,
			expected: []*TOCItem{
				{
					Level: 2,
					Id:    "h2",
					Title: "Heading 2",
					Children: []*TOCItem{
						{
							Level: 3,
							Id:    "h3",
							Title: "Heading 3",
						},
					},
				},
			},
			error: nil,
		},
	}

	for _, tc := range testCases {
		ret, err := ParseHTMLToTOCWithDepthRange(bytes.NewBuffer([]byte(tc.html)), tc.minDepth, tc.maxDepth)
		assert.Equal(t, tc.error, err)
		assert.Equal(t, tc.expected, ret)
	}
}
