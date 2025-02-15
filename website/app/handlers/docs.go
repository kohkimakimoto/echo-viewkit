package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kohkimakimoto/echo-viewkit/website/app/view"
	"github.com/labstack/echo/v4"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"io/fs"
	"net/http"
	"strings"
)

type DocsProps struct {
	Title   string          `pongo2:"title"`
	Toc     []*view.TOCItem `pongo2:"toc"`
	Content string          `pongo2:"content"`
}

func DocsHandler(f fs.FS) echo.HandlerFunc {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
			emoji.Emoji,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return func(c echo.Context) error {
		name := strings.TrimPrefix(c.Param("*"), "/")
		if name == "" {
			name = "index"
		}
		filePath := name + ".md"

		// Load the markdown file
		mdContent, err := fs.ReadFile(f, filePath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return echo.ErrNotFound
			}
			return fmt.Errorf("failed to read markdown file %q: %w", filePath, err)
		}

		// Convert the markdown to HTML
		var buf bytes.Buffer
		context := parser.NewContext()
		if err := md.Convert(mdContent, &buf, parser.WithContext(context)); err != nil {
			return fmt.Errorf("failed to convert markdown: %w", err)
		}
		content := buf.String()

		// Get the metadata
		metaData := meta.Get(context)
		var title string
		if v, ok := metaData["title"].(string); ok {
			title = v
		}

		// Get the table of contents
		toc, err := view.ParseHTMLToTOCWithDepthRange(strings.NewReader(content), 2, 3)
		if err != nil {
			return fmt.Errorf("failed to parse HTML to TOC: %w", err)
		}

		return c.Render(http.StatusOK, "pages/docs", &DocsProps{
			Title:   title,
			Toc:     toc,
			Content: content,
		})
	}
}
