package pongo2

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexRemove_Execute(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "remove custom tags",
			patterns: []string{`(?s)<!--\s*DEBUG\s*-->.*?<!--\s*/DEBUG\s*-->`},
			input: `
<div>visible</div>
<!-- DEBUG -->
debug info
<!-- /DEBUG -->
<div>visible2</div>
`,
			expected: `
<div>visible</div>

<div>visible2</div>
`,
			hasError: false,
		},
		{
			name:     "no match keeps input as is",
			patterns: []string{`(?s)<remove>.*?</remove>`},
			input: `
<div>
content
</div>
`,
			expected: `
<div>
content
</div>
`,
			hasError: false,
		},
		{
			name:     "invalid regex returns error",
			patterns: []string{`[`},
			input:    "",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := NewRegexRemove(tt.patterns...)

			if tt.hasError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			src := strings.NewReader(tt.input)
			dst := &bytes.Buffer{}

			err = processor.Execute(dst, src)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, dst.String())
		})
	}
}

func TestMustNewRegexRemove(t *testing.T) {
	t.Run("valid patterns", func(t *testing.T) {
		processor := MustNewRegexRemove(`(?s)<style[^>]*>.*?</style>`, `(?s)<script[^>]*>.*?</script>`)
		assert.NotNil(t, processor)

		input := `<style>test</style><script>test</script><div>content</div>`
		src := strings.NewReader(input)
		dst := &bytes.Buffer{}

		err := processor.Execute(dst, src)

		assert.NoError(t, err)
		assert.Equal(t, "<div>content</div>", dst.String())
	})

	t.Run("invalid pattern panics", func(t *testing.T) {
		assert.Panics(t, func() {
			MustNewRegexRemove(`[`)
		})
	})
}

func TestRegexRemove_TemplateExtractEquivalent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "remove <style data-extract> tag",
			input: `
<style data-extract>
.test { color: red; }
</style>
<div>content</div>
`,
			expected: `

<div>content</div>
`,
		},
		{
			name: "keep style tag without data-extract attribute",
			input: `
<style>
.normal { color: black; }
</style>
`,
			expected: `
<style>
.normal { color: black; }
</style>
`,
		},
		{
			name: "remove script tag with data-extract attribute",
			input: `<script data-extract>
...
</script>
<div>
...
</div>`,
			expected: `
<div>
...
</div>`,
		},
		{
			name: "keep normal script tag",
			input: `<script>
console.log('normal script');
</script>
<div>Content</div>`,
			expected: `<script>
console.log('normal script');
</script>
<div>Content</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := MustNewRegexRemove(
				`(?i)(?s)<style[^>]*\bdata-extract\b[^>]*>.*?</style>`,
				`(?i)(?s)<script[^>]*\bdata-extract\b[^>]*>.*?</script>`,
			)
			src := strings.NewReader(tt.input)
			dst := &bytes.Buffer{}

			err := processor.Execute(dst, src)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, dst.String())
		})
	}
}
