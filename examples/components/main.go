package main

import (
	"net/http"

	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	v := viewkit.New()
	v.Debug = true
	v.BaseDir = "views"
	v.Components = []*pongo2.Component{
		Alert,
		Greeting,
	}
	v.InlineComponents = []*pongo2.InlineComponent{
		InlineAlert,
	}
	v.AnonymousComponents = []*pongo2.AnonymousComponent{
		{Name: "primary-button", TemplateFile: "components/primary-button"},
	}
	v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
		{Dir: "components"},
	}
	e.Renderer = v.MustRenderer()
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", nil)
	})
	e.Logger.Fatal(e.Start(":1323"))
}

var Greeting = &pongo2.Component{
	Name:         "greeting",
	TemplateFile: "components/greeting",
	Props:        []string{"color", "message"},
	Setup: func(ctx *pongo2.ComponentExecutionContext) error {
		ctx.Default("color", "blue")
		ctx.Default("message", "Hello, World!")
		return nil
	},
}

type AlertData struct {
	Message string `pongo2:"message"`
	Type    string `pongo2:"type"`
}

var Alert = &pongo2.Component{
	Name:         "alert",
	TemplateFile: "components/alert",
	Props:        []string{"message", "type"},
	Setup: func(ctx *pongo2.ComponentExecutionContext) error {
		if err := ctx.Defaults(&AlertData{
			Message: "This is an alert message.",
			Type:    "info",
		}); err != nil {
			return err
		}
		return nil
	},
}

var InlineAlert = &pongo2.InlineComponent{
	Name:  "alert",
	Props: []string{"message", "type"},
	TemplateString: `
<div class="alert alert-{{ type }}">
	{{ message }}
</div>
`,
}
