package main

import (
	"net/http"

	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	v := viewkit.New()
	v.BaseDir = "views"
	e.Renderer = v.MustRenderer()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", map[string]any{
			"name": "Echo ViewKit",
		})
	})
	e.Logger.Fatal(e.Start(":1323"))
}
