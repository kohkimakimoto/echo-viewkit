package viewkit

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ViewHandler returns a handler that renders response using a view.
// The name argument is the name of the view to render.
func ViewHandler(name string) echo.HandlerFunc {
	return ViewHandlerWithData(name, nil)
}

// ViewHandlerWithData returns a handler that renders response using a view.
// The name argument is the name of the view to render.
// The data argument is the data to the view.
func ViewHandlerWithData(name string, data any) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, name, data)
	}
}
