---
title: Getting Started
---

# Getting Started

This section provides a brief guide on getting started with Echo ViewKit,
including step-by-step instructions for installing Echo ViewKit and setting up a simple web application.

This guide assumes that you are familiar with the basics of Echo application development.
If you are new to Echo, please read the [Echo documentation](https://echo.labstack.com/) first.

## Installation

To install Echo ViewKit, run the following command in your Go project:

```shell
go get github.com/kohkimakimoto/echo-viewkit
```

## Create a view template

In the root directory of your Go project, create a new directory named `views`.

```shell
mkdir views
```

In the `views` directory, create a new file named `index.html` with the following content:

```html
<html>
  <body>
    <h1>Hello {{ name }}!</h1>
  </body>
</html>
```

## Create a simple web application

Create a new Go file named `main.go` in the root directory of your Go project with the following content:

```go
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
```

## Start the server

Run the following command to start the server:

```shell
go run main.go
```

Browse to [http://localhost:1323](http://localhost:1323) to see the rendered template.
