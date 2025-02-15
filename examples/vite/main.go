package main

import (
	"flag"
	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/labstack/echo/v4"
	"net/http"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	e := echo.New()

	v := viewkit.New()
	v.Debug = debug
	v.BaseDir = "views"
	// Enable Vite integration
	v.Vite = true
	v.ViteDevMode = v.Debug
	if !v.ViteDevMode {
		// vite production build config
		v.ViteManifest = viewkit.MustParseViteManifestFile("public/build/manifest.json")
		v.ViteBasePath = "/build"
	}

	e.Renderer = v.MustRenderer()
	e.Static("/", "public")
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", nil)
	})

	if v.ViteDevMode {
		// Start Vite dev server if it's in dev mode
		go func() {
			if err := v.StartViteDevServer(); err != nil {
				e.Logger.Errorf("the vite dev server returned an error: %v", err)
			}
		}()
	}

	e.Logger.Fatal(e.Start(":1323"))
}
