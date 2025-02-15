package main

import (
	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/labstack/echo/v4"
	"net/http"
)

func main() {
	e := echo.New()

	v := viewkit.New()
	v.BaseDir = "views"
	//v.SharedContextProviders = map[string]viewkit.SharedContextProviderFunc{
	//	"siteName": func(c echo.Context) (any, error) {
	//		return "Echo ViewKit Example", nil
	//	},
	//}
	v.SharedContextProviders = map[string]viewkit.SharedContextProviderFunc{
		"siteName": func(c echo.Context) (any, error) {
			return func() string {
				return "Echo ViewKit Example"
			}, nil
		},
	}

	e.Renderer = v.MustRenderer()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", nil)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
