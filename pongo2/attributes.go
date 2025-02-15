package pongo2

import (
	"fmt"
	"html"
	"strings"
)

// Attributes represents a collection of HTML attributes with order preserved.
type Attributes struct {
	attrs map[string]string
	order []string
}

// newAttributes initializes a new Attributes instance.
func newAttributes(pairs [][2]string) *Attributes {
	attrs := make(map[string]string)
	order := make([]string, 0, len(pairs))

	for _, pair := range pairs {
		key, value := pair[0], pair[1]
		// Skip duplicate keys, keep the first occurrence
		if _, exists := attrs[key]; !exists {
			order = append(order, key)
		}
		attrs[key] = value
	}

	return &Attributes{
		attrs: attrs,
		order: order,
	}
}

func (a *Attributes) Len() int {
	return len(a.attrs)
}

// String generates the full string representation of the attributes.
// Note:
// Do not call this function directly in templates like {{ attributes.String() }}.
// Instead, use {{ attributes }}.
// Since the String method implements the fmt.Stringer interface,
// Pongo2 automatically calls this method internally to generate a raw string representation.
// For this reason, the return value of this method is not escaped.
// Calling this method directly can lead to security vulnerabilities.
func (a *Attributes) String() string {
	var parts []string
	for _, key := range a.order {
		if value, exists := a.attrs[key]; exists {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, html.EscapeString(key), html.EscapeString(value)))
		}
	}
	return strings.Join(parts, " ")
}

// Only extracts a subset of attributes by the specified keys.
func (a *Attributes) Only(keys ...string) *Attributes {
	newAttrs := make(map[string]string)
	newOrder := make([]string, 0)
	for _, key := range keys {
		if value, ok := a.attrs[key]; ok {
			newAttrs[key] = value
			newOrder = append(newOrder, key)
		}
	}
	return &Attributes{attrs: newAttrs, order: newOrder}
}

// Without generates a subset of attributes excluding the specified keys.
func (a *Attributes) Without(keys ...string) *Attributes {
	newAttrs := make(map[string]string)
	newOrder := make([]string, 0)
	exclude := make(map[string]struct{})
	for _, key := range keys {
		exclude[key] = struct{}{}
	}
	for _, key := range a.order {
		if _, found := exclude[key]; !found {
			if value, ok := a.attrs[key]; ok {
				newAttrs[key] = value
				newOrder = append(newOrder, key)
			}
		}
	}
	return &Attributes{attrs: newAttrs, order: newOrder}
}

// Get retrieves the value of a specific attribute.
func (a *Attributes) Get(key string) string {
	return a.attrs[key]
}

// Has checks if a specific attribute exists.
func (a *Attributes) Has(key string) bool {
	_, exists := a.attrs[key]
	return exists
}

// Default sets the default value for a single attribute if it doesn't already exist.
func (a *Attributes) Default(key, value string) *Attributes {
	newAttrs := make(map[string]string)
	newOrder := append([]string{}, a.order...)

	// Copy existing attributes
	for k, v := range a.attrs {
		newAttrs[k] = v
	}

	if key == "class" {
		// Prepend the default class value
		if existing, ok := newAttrs[key]; ok {
			newAttrs[key] = value + " " + existing
		} else {
			newAttrs[key] = value
			newOrder = append(newOrder, key)
		}
	} else {
		// Set default value if the key does not exist
		if _, exists := newAttrs[key]; !exists {
			newAttrs[key] = value
			newOrder = append(newOrder, key)
		}
	}

	return &Attributes{attrs: newAttrs, order: newOrder}
}
