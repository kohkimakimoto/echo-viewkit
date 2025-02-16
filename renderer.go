package viewkit

import (
	"bytes"
	"io"
	"strings"

	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/labstack/echo/v4"
)

// Renderer is a renderer implementation for Echo.
// see https://echo.labstack.com/docs/templates
type Renderer struct {
	templateSet *pongo2.TemplateSet
	providers   map[string]SharedContextProviderFunc
}

func (r *Renderer) TemplateSet() *pongo2.TemplateSet {
	return r.templateSet
}

func (r *Renderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	// check the fragment.
	// If the name has '#', it means the template name with the fragment.
	templateName, fragmentName := parseFragment(name)
	t, err := r.templateSet.FromCache(templateName)
	if err != nil {
		return err
	}
	pongo2Context, err := pongo2.MarshalContext(data)
	if err != nil {
		return err
	}

	for k, provider := range r.providers {
		v, err := provider(c)
		if err != nil {
			return err
		}
		pongo2Context[k] = v
	}

	if fragmentName != "" {
		return t.ExecuteFragmentWriterWithEchoContext(pongo2Context, fragmentName, w, c)
	} else {
		return t.ExecuteWriterWithEchoContext(pongo2Context, w, c)
	}
}

// SharedContextProviderFunc is a function that provides shared context data
type SharedContextProviderFunc func(c echo.Context) (any, error)

// Render is a helper function to render a template with the specified renderer.
// It is useful when you want to render a template with a renderer other than the default renderer.
func Render(renderer echo.Renderer, c echo.Context, code int, name string, data any) (err error) {
	buf := new(bytes.Buffer)
	if err = renderer.Render(buf, name, data, c); err != nil {
		return
	}
	return c.HTMLBlob(code, buf.Bytes())
}

func parseFragment(s string) (templateName string, fragmentName string) {
	if idx := strings.Index(s, "#"); idx != -1 {
		return s[:idx], s[idx+1:]
	}
	return s, ""
}
