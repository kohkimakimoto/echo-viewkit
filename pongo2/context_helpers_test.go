package pongo2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshalContext(t *testing.T) {
	data := &struct {
		Name string `pongo2:"name"`
		Age  int    `pongo2:"age"`
	}{
		Name: "Alice",
		Age:  20,
	}

	c, err := MarshalContext(data)
	if err != nil {
		t.Error(err)
	}
	if c["name"] != "Alice" {
		t.Errorf("Expected 'Alice', got %v", c["name"])
	}
	if c["age"] != 20 {
		t.Errorf("Expected 20, got %v", c["age"])
	}
}

func TestUnmarshalContext(t *testing.T) {
	t.Run("unmarshal private struct fields", func(t *testing.T) {
		c := Context{
			"str_val":   AsValue("aaaa"),
			"int_val":   AsValue(20),
			"int64_val": AsValue(int64(30)),
		}

		data := &struct {
			Str   string `pongo2:"str_val"`
			Int   int    `pongo2:"int_val"`
			Int64 int64  `pongo2:"int64_val"`
		}{}

		err := UnmarshalContext(c, data)
		assert.NoError(t, err)
		assert.Equal(t, "aaaa", data.Str)
		assert.Equal(t, 20, data.Int)
		assert.Equal(t, int64(30), data.Int64)
	})

	t.Run("unmarshal struct struct fields", func(t *testing.T) {
		type Tag struct {
			Name string
		}

		c := Context{
			"tag": AsValue(&Tag{
				Name: "tag",
			}),
		}

		data := &struct {
			Tag *Tag `pongo2:"tag"`
		}{}

		err := UnmarshalContext(c, data)
		assert.NoError(t, err)
		assert.Equal(t, "tag", data.Tag.Name)
	})
}
