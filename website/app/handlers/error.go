package handlers

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type HTTPErrorProps struct {
	Title   string `pongo2:"title"`
	Code    int    `pongo2:"code"`
	Message string `pongo2:"message"`
}

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// try to convert error to HTTP error
	var he *echo.HTTPError
	if !errors.As(err, &he) {
		// not HTTP error, create internal server error
		he = &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  http.StatusText(http.StatusInternalServerError),
			Internal: err,
		}
	}

	if he.Code >= 500 {
		c.Logger().Errorf("%+v", err)
	}

	if c.Request().Method == echo.HEAD {
		// see https://github.com/labstack/echo/issues/608
		_ = c.NoContent(he.Code)
		return
	}

	_ = c.Render(he.Code, "pages/error", &HTTPErrorProps{
		Title:   fmt.Sprintf("%d %s | Echo ViewKit", he.Code, http.StatusText(he.Code)),
		Code:    he.Code,
		Message: http.StatusText(he.Code),
	})
}
