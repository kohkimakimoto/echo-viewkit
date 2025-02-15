package viewkit

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"net/url"
)

func IsDebugFunctionProvider(debug bool) SharedContextProviderFunc {
	return func(c echo.Context) (any, error) {
		return func() bool {
			return debug
		}, nil
	}
}

func URLPathFunctionProvider() SharedContextProviderFunc {
	return func(c echo.Context) (any, error) {
		return func() string {
			return c.Request().URL.Path
		}, nil
	}
}

func URLQueryFunctionProvider() SharedContextProviderFunc {
	return func(c echo.Context) (any, error) {
		return func(key ...string) (string, error) {
			if len(key) == 1 {
				return c.QueryParam(key[0]), nil
			} else if len(key) > 1 {
				return "", errors.New("too many arguments")
			}
			return c.Request().URL.RawQuery, nil
		}, nil
	}
}

func URLPathQueryFunctionProvider() SharedContextProviderFunc {
	return func(c echo.Context) (any, error) {
		return func(keys ...string) string {
			req := c.Request()
			if len(keys) == 0 {
				return req.URL.RequestURI()
			}

			basePath := req.URL.Path

			var queryString string
			for i, key := range keys {
				if value := c.QueryParam(key); value != "" {
					if i > 0 {
						queryString += "&"
					}
					queryString += key + "=" + url.QueryEscape(value)
				}
			}

			if queryString == "" {
				return basePath
			}

			return basePath + "?" + queryString
		}, nil
	}
}

func JsonMarshalFunctionProvider() SharedContextProviderFunc {
	return func(c echo.Context) (any, error) {
		return func(v any) (string, error) {
			j, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(j), nil
		}, nil
	}
}
