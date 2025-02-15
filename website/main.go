package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/kohkimakimoto/echo-viewkit/website/app/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	//go:embed resources/views
	embeddedViewsFS embed.FS
	//go:embed resources/docs
	embeddedDocsFS embed.FS
	//go:embed public
	embeddedPublicFS embed.FS
)

func main() {
	if err := realMain(); err != nil {
		log.Fatal(err)
	}
}

func realMain() error {
	//------------------------------------------------------------------------------------------
	// Parse the command line options and init parameters
	//------------------------------------------------------------------------------------------
	var viewsFS = echo.MustSubFS(embeddedViewsFS, "resources/views")
	var docsFS = echo.MustSubFS(embeddedDocsFS, "resources/docs")
	var publicFS = echo.MustSubFS(embeddedPublicFS, "public")
	var port string
	var debug bool

	// Options
	flag.StringVar(&port, "port", "8080", "Listen port. Default is 8080. You can also use PORT environment variable.")
	flag.BoolVar(&debug, "debug", false, "Debug mode. Default is false. You can also use DEBUG environment variable.")

	flag.VisitAll(func(f *flag.Flag) {
		// Set the flag value from the environment variable if the variable exists.
		name := f.Name
		if s := os.Getenv(strings.Replace(strings.ToUpper(name), "-", "_", -1)); s != "" {
			_ = f.Value.Set(s)
		}
	})

	flag.Usage = func() {
		fmt.Print(`Usage: echo-viewkit-website [OPTIONS...]

The echo-viewkit website.

Options:
  -port N           Port number to listen.
  -debug            Run on a debug mode.
  -h, -help         Show help.
`)
	}
	flag.Parse()

	if debug {
		// In debug mode, the current working directory must be the "website" directory,
		// because the app uses directories relative to the current working directory.

		// Set FS instances to the local file system instead of the embedded file system
		// because embed.FS does not support hot reloading.

		// view
		viewsFS = os.DirFS("resources/views")
		// docs
		docsFS = os.DirFS("resources/docs")
		// public
		publicFS = os.DirFS("public")
	}

	//------------------------------------------------------------------------------------------
	// Init the server app
	//------------------------------------------------------------------------------------------
	e := echo.New()
	e.Debug = debug
	e.HTTPErrorHandler = handlers.HTTPErrorHandler

	// Init Echo ViewKit instance
	v := viewkit.New()
	v.Debug = debug
	v.FS = viewsFS
	v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
		{Dir: "components"},
	}
	v.Vite = true
	v.ViteDevMode = debug

	if !v.ViteDevMode {
		// vite production build config
		v.ViteManifest = viewkit.MustParseViteManifestFS(publicFS, "build/manifest.json")
		v.ViteBasePath = "/build"
	}

	// Set the view renderer
	e.Renderer = v.MustRenderer()

	// Middleware
	e.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusFound,
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Static files
	e.StaticFS("/", publicFS)

	// Handlers
	e.GET("/", viewkit.ViewHandler("pages/index"))
	e.GET("/docs*", handlers.DocsHandler(docsFS))

	if v.ViteDevMode {
		// Start Vite dev server if it's in dev mode
		go func() {
			if err := v.StartViteDevServer(); err != nil {
				e.Logger.Errorf("the vite dev server returned an error: %v", err)
			}
		}()
	}

	// Start the server
	if err := e.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
