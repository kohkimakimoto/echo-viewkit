package view

import (
	"golang.org/x/net/html"
	"io"
	"strings"
)

// TOCItem represents an item in the table of contents.
type TOCItem struct {
	Level    int    // Heading level (h1 -> 1, h2 -> 2, etc.)
	Id       string // Value of the "id" attribute
	Title    string // Text content of the heading
	Children []*TOCItem
}

// ParseHTMLToTOC parses HTML from an io.Reader and builds a tree of TOCItems
// with no minimum or maximum depth limits (same as before).
func ParseHTMLToTOC(r io.Reader) ([]*TOCItem, error) {
	return ParseHTMLToTOCWithDepthRange(r, 0, 0)
}

// ParseHTMLToTOCWithDepthRange parses HTML from an io.Reader and
// builds a tree of TOCItems only for headings whose level is in [minDepth, maxDepth].
//
// - If minDepth = 0, there is no lower bound. (i.e. h1 is not skipped by default)
// - If maxDepth = 0, there is no upper bound. (i.e. h6 is not skipped by default)
//
// Examples:
//   - minDepth=2, maxDepth=0 -> skip h1, keep h2~h6
//   - minDepth=2, maxDepth=3 -> skip h1, skip h4~h6, keep h2 and h3
//   - minDepth=0, maxDepth=2 -> skip h3~h6, keep h1 and h2
func ParseHTMLToTOCWithDepthRange(r io.Reader, minDepth, maxDepth int) ([]*TOCItem, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Extract a flat list of headings (h1~h6)
	headings := extractHeadings(doc)

	// Build a tree-structured TOC from the flat list of headings,
	// applying the depth restrictions.
	toc := buildTOC(headings, minDepth, maxDepth)
	return toc, nil
}

// extractHeadings traverses the parsed DOM tree and extracts
// h1~h6 elements into a flat slice of TOCItem.
func extractHeadings(n *html.Node) []TOCItem {
	var headings []TOCItem

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			level := headingLevel(n.Data)
			if level != 0 {
				// Retrieve the "id" attribute
				id := ""
				for _, attr := range n.Attr {
					if attr.Key == "id" {
						id = attr.Val
						break
					}
				}
				// Extract text content from the heading
				title := getTextContent(n)

				headings = append(headings, TOCItem{
					Level: level,
					Id:    id,
					Title: title,
				})
			}
		}
		// Recursively traverse child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return headings
}

// headingLevel converts a heading tag name (e.g. "h2") to its numeric level.
// Returns 0 if it's not h1~h6.
func headingLevel(tag string) int {
	switch tag {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		return 0
	}
}

// getTextContent retrieves concatenated text content from the given node and
// all its descendants.
func getTextContent(n *html.Node) string {
	var sb strings.Builder

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}

// buildTOC converts a flat list of TOCItem into a tree-structured TOC by
// mapping items to their respective parents according to the heading level.
//
// If minDepth > 0, headings with level < minDepth are skipped.
// If maxDepth > 0, headings with level > maxDepth are skipped.
// If both are 0, there is no limit in either direction.
func buildTOC(headings []TOCItem, minDepth, maxDepth int) []*TOCItem {
	if len(headings) == 0 {
		return nil
	}

	var root []*TOCItem
	var stack []*TOCItem

	for _, h := range headings {
		// Check minDepth
		if minDepth > 0 && h.Level < minDepth {
			continue
		}
		// Check maxDepth
		if maxDepth > 0 && h.Level > maxDepth {
			continue
		}

		item := &TOCItem{
			Level: h.Level,
			Id:    h.Id,
			Title: h.Title,
		}

		// Pop from the stack if the current level is <= the level on top of the stack
		for len(stack) > 0 && stack[len(stack)-1].Level >= item.Level {
			stack = stack[:len(stack)-1]
		}

		// If stack is empty, it means this item is top-level
		if len(stack) == 0 {
			root = append(root, item)
		} else {
			// Otherwise, the top item in stack is the parent
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, item)
		}

		// Push the current item to the stack
		stack = append(stack, item)
	}

	return root
}
