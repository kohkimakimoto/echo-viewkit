package pongo2

import (
	"io"
	"regexp"
)

// RegexRemove is a preprocessor that removes parts matching regular expressions
type RegexRemove struct {
	patterns []*regexp.Regexp
}

// NewRegexRemove creates a new instance of RegexRemove
// patterns can specify multiple regular expression pattern strings to match parts to be removed
func NewRegexRemove(patterns ...string) (*RegexRemove, error) {
	regexps := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		regexps[i] = re
	}
	return &RegexRemove{patterns: regexps}, nil
}

// MustNewRegexRemove creates a new instance of RegexRemove and panics if any pattern is invalid
// This is useful when patterns are statically determined and error handling can be simplified
func MustNewRegexRemove(patterns ...string) *RegexRemove {
	p, err := NewRegexRemove(patterns...)
	if err != nil {
		panic(err)
	}
	return p
}

func (p *RegexRemove) Execute(dst io.Writer, src io.Reader) error {
	b, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	// Remove parts based on each regular expression pattern in order
	for _, re := range p.patterns {
		b = re.ReplaceAll(b, []byte(""))
	}

	_, err = dst.Write(b)
	return err
}
