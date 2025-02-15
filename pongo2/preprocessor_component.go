package pongo2

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type ComponentHTMLTagPreProcessorConfig struct {
	// Prefix for custom tags (e.g., "x-").
	TagPrefix string
}

type HTMLAttribute struct {
	Name  string
	Value string
}

var (
	// Regex pattern to match {% verbatim %}...{% endverbatim %} tags
	verbatimRegex            = regexp.MustCompile(`(?s){% verbatim %}.*?{% endverbatim %}`)
	verbatimPlaceholderRegex = regexp.MustCompile(`##VERBATIM_BLOCK_(\d+)##`)
)

func ComponentHTMLTagPreProcessor(config ComponentHTMLTagPreProcessorConfig) PreProcessorFunc {
	// Regex pattern to match opening and closing tags (e.g., <x-alert>...</x-alert>)
	openingComponentTagRegex := regexp.MustCompile(fmt.Sprintf(`<%s([a-zA-Z0-9_.-]+)([^<>]*?)\s*>`, regexp.QuoteMeta(config.TagPrefix)))
	closingComponentTagRegex := regexp.MustCompile(fmt.Sprintf(`</%s([a-zA-Z0-9_.-]+)>`, regexp.QuoteMeta(config.TagPrefix)))

	// Regex pattern to match self-closing tags (e.g., <x-alert />)
	selfClosingRegex := regexp.MustCompile(fmt.Sprintf(`<%s([a-zA-Z0-9_.-]+)([^<>]*?)\s*/>`, regexp.QuoteMeta(config.TagPrefix)))

	// Regex pattern to match opening and closing <x-slot> tags
	openingNamedSlotTagRegex := regexp.MustCompile(fmt.Sprintf(`<%sslot([^<>]*?)\s*>`, regexp.QuoteMeta(config.TagPrefix)))
	closingSlotTagRegex := regexp.MustCompile(fmt.Sprintf(`</%sslot>`, regexp.QuoteMeta(config.TagPrefix)))

	return func(dst io.Writer, src io.Reader) error {
		// Read input data
		b, err := io.ReadAll(src)
		if err != nil {
			return err
		}
		txt := string(b)

		var verbatimPlaceholders []string
		var verbatimPlaceholderIndex int

		// Replace {% verbatim %}...{% endverbatim %} tags with placeholders.
		// Because we must not process the content inside the verbatim tags.
		txt = verbatimRegex.ReplaceAllStringFunc(txt, func(m string) string {
			verbatimPlaceholders = append(verbatimPlaceholders, m)
			ph := fmt.Sprintf("##VERBATIM_BLOCK_%d##", verbatimPlaceholderIndex)
			verbatimPlaceholderIndex++
			return ph
		})

		// Replace <x-slot> tags
		txt = openingNamedSlotTagRegex.ReplaceAllStringFunc(txt, func(match string) string {
			matches := openingNamedSlotTagRegex.FindStringSubmatch(match)
			if len(matches) < 2 {
				return match // Return as is if the match is invalid
			}

			attributes := strings.TrimSpace(matches[1]) // Attributes section

			// Parse attributes
			orderedAttrs := parseComponentAttributes(attributes)

			// Generate the slot tag
			var buffer bytes.Buffer
			slotName := ""
			for _, attr := range orderedAttrs {
				if convertToCamelCase(attr.Name) == "name" {
					slotName = attr.Value
					break
				}
			}

			if slotName == "" {
				return "{% slot %}"
			}

			buffer.WriteString(fmt.Sprintf(`{%% slot "%s" %%}`, slotName))
			return buffer.String()
		})

		txt = closingSlotTagRegex.ReplaceAllStringFunc(txt, func(match string) string {
			return "{% endslot %}"
		})

		// Replace self-closing component tags
		txt = selfClosingRegex.ReplaceAllStringFunc(txt, func(match string) string {
			matches := selfClosingRegex.FindStringSubmatch(match)
			if len(matches) < 3 {
				return match // Return as is if the match is invalid
			}

			//componentName := strings.ReplaceAll(matches[1], ".", "/") // Tag name generates component name (e.g., "alert" => "alert", "namespace.alert" => "namespace/alert")
			componentName := matches[1]                 // Component name (e.g., "alert" "layouts.app")
			attributes := strings.TrimSpace(matches[2]) // Attributes section

			// Parse attributes and preserve order
			orderedAttrs := parseComponentAttributes(attributes)

			// Generate the template tag
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf(`{%% component "%s"`, componentName))

			// Add attributes in the original order
			slotData := ""
			normalAttrs := []string{}
			for _, attr := range orderedAttrs {
				if attr.Name == "slot-data" {
					slotData = fmt.Sprintf(` slotData="%s"`, attr.Value)
				} else if strings.HasPrefix(attr.Name, ":") {
					normalAttrs = append(normalAttrs, fmt.Sprintf(` "%s"=%s`, strings.TrimPrefix(attr.Name, ":"), attr.Value))
				} else {
					normalAttrs = append(normalAttrs, fmt.Sprintf(` "%s"="%s"`, attr.Name, attr.Value))
				}
			}

			// Add slotData if exists
			if slotData != "" {
				buffer.WriteString(slotData)
			}

			// Add remaining attributes
			if len(normalAttrs) > 0 {
				buffer.WriteString(" withAttrs")
				for _, attr := range normalAttrs {
					buffer.WriteString(attr)
				}
			}

			buffer.WriteString(" %}{% endcomponent %}")
			return buffer.String()
		})

		// Replace component tags
		txt = openingComponentTagRegex.ReplaceAllStringFunc(txt, func(match string) string {
			matches := openingComponentTagRegex.FindStringSubmatch(match)
			if len(matches) < 3 {
				return match // Return as is if the match is invalid
			}

			componentName := matches[1]                 // Component name (e.g., "alert" "layouts.app")
			attributes := strings.TrimSpace(matches[2]) // Attributes section

			// Parse attributes and preserve order
			orderedAttrs := parseComponentAttributes(attributes)

			// Generate the template tag
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf(`{%% component "%s"`, componentName))

			// Add attributes in the original order
			slotData := ""
			normalAttrs := []string{}
			for _, attr := range orderedAttrs {
				if attr.Name == "slot-data" {
					slotData = fmt.Sprintf(` slotData="%s"`, attr.Value)
				} else if strings.HasPrefix(attr.Name, ":") {
					normalAttrs = append(normalAttrs, fmt.Sprintf(` "%s"=%s`, strings.TrimPrefix(attr.Name, ":"), attr.Value))
				} else {
					normalAttrs = append(normalAttrs, fmt.Sprintf(` "%s"="%s"`, attr.Name, attr.Value))
				}
			}

			// Add slotData if exists
			if slotData != "" {
				buffer.WriteString(slotData)
			}

			// Add remaining attributes
			if len(normalAttrs) > 0 {
				buffer.WriteString(" withAttrs")
				for _, attr := range normalAttrs {
					buffer.WriteString(attr)
				}
			}

			buffer.WriteString(" %}")
			return buffer.String()
		})

		txt = closingComponentTagRegex.ReplaceAllStringFunc(txt, func(match string) string {
			matches := closingComponentTagRegex.FindStringSubmatch(match)
			if len(matches) < 2 {
				return match // Return as is if the match is invalid
			}
			return fmt.Sprintf("{%% endcomponent %%}")
		})

		// Replace the placeholders with the original verbatim blocks
		txt = verbatimPlaceholderRegex.ReplaceAllStringFunc(txt, func(m string) string {
			subMatches := verbatimPlaceholderRegex.FindStringSubmatch(m)
			if len(subMatches) < 2 {
				return m
			}
			idxStr := subMatches[1]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return m
			}
			// Get the original verbatim block
			return verbatimPlaceholders[idx]
		})

		// Write the converted text to the output
		if _, err := dst.Write([]byte(txt)); err != nil {
			return err
		}
		return nil
	}
}

// parseComponentAttributes parses the HTML attribute string into a slice of Attribute preserving order.
func parseComponentAttributes(attr string) []HTMLAttribute {
	// Use regex to extract attributes
	attrRegex := regexp.MustCompile(`([a-zA-Z0-9_:.-]+)\s*=\s*"(.*?)"`)
	matches := attrRegex.FindAllStringSubmatch(attr, -1)

	var attrs []HTMLAttribute
	for _, match := range matches {
		if len(match) >= 3 {
			attrs = append(attrs, HTMLAttribute{
				Name:  match[1],
				Value: match[2],
			})
		}
	}
	return attrs
}
