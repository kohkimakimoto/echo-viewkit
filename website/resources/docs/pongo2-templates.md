---
title: Pongo2 Templates
---

# Pongo2 Templates

[Pongo2](https://github.com/flosch/pongo2) is a Django-like syntax template engine for Go.
It was created by [Florian Schlachter](https://github.com/flosch).
Echo ViewKit uses a forked version of Pongo2, extensively customizing it to meet the demands of modern web development and seamless Echo integration.

## Template syntax

Pongo2 syntax is similar to the Django template language.
Currently, we do not cover Pongo2 syntax in this documentation.
However, you can find detailed specifications at the following links:

- [Pongo2](https://github.com/flosch/pongo2)
- [The Django Template Language](https://django.readthedocs.io/en/1.7.x/topics/templates.html)

## View renderer

Echo ViewKit provides an Echo [renderer](https://pkg.go.dev/github.com/labstack/echo#Renderer) implementation integrated with the Pongo2 template engine.
The renderer is an interface that allows you to render templates in Echo applications.
Here is a simple example of how to use the renderer:

```go
// Create a new Echo instance.
e := echo.New()
// Create a new Echo ViewKit instance.
v := viewkit.New()
// Set the base directory of the template files.
v.BaseDir = "views"
// Create a new renderer and set it to the Echo instance.
e.Renderer = v.MustRenderer()
```

After setting up the renderer, you can call the `Render` method in handler functions to render a template:

```go
func HelloHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "hello", map[string]any{
		"message": "Hello World",
	})
}
```

The first argument of the `Render` method is the HTTP status code,
the second argument is the template file name (You can omit the `.html` file extension),
and the third argument is the data to pass to the template.

## Debug mode

When the view renderer renders the templates, they are compiled and cached by default.
This improves rendering performance but it can be inconvenient during development because updates to template files are not immediately reflected.
To disable caching, set the `Debug` property of the ViewKit instance to `true`:

```go
v := viewkit.New()
v.Debug = true
```

## Passing data to templates

As you saw in the previous examples, you can pass data to the template by providing a `map[string]any` map.

```go
c.Render(http.StatusOK, "hello", map[string]any{
	"message": "Hello World",
})
```

This data can be accessed in the template using the key names.

```html
<h1>{{ message }}</h1>
```

If you prefer to use typed values instead of a map, you can use a struct with the `pongo2` tag.

```go
type HelloProps struct {
	Message string `pongo2:"message"`
}

func HelloHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "hello", &HelloProps{
		Message: "Hello World",
	})
}
```

## Template location configuration

The simplest configuration of the view renderer is just to set the `BaseDir` property.
It is the directory path where the template files are located.

```go
v := viewkit.New()
v.BaseDir = "views"
```

You can also load templates from Go `fs.FS` interface instead of the directory path.
It is useful when you want to embed template files into a binary.
Echo ViewKit provides the `FS` property to set the `fs.FS` interface.

```go
import (
	"embed"
	viewkit "github.com/kohkimakimoto/echo-viewkit"
)

//go:embed views
var viewFS embed.FS

func main() {
	v := viewkit.New()
	v.FS = viewFS
	v.FSBaseDir = "views"
	// ...
}
```

## Shared context

Sometimes you want to access the same data or functions in all templates.
The shared context is useful for these cases.

> :memo:
> As explained in the [Components](/docs/components) section,
> encapsulated components have their own context, meaning that data or functions from the parent context are not accessible.
> The shared context provides a mechanism to share data or functions across all templates, including components.


### Shared context providers

To use the shared context, you may set the `SharedContextProviders` property.

```go
v := viewkit.New()
v.SharedContextProviders = map[string]viewkit.SharedContextProviderFunc{
	"siteName": func(c echo.Context) (any, error) {
		return "Echo ViewKit Example", nil
	},
}
```

As you can see in the example above, `SharedContextProviders` is a map of `string` and `SharedContextProviderFunc`.
The `SharedContextProviderFunc` is a function that returns shared data to be passed to all templates when rendering.
This function is called whenever the renderer renders a template, ensuring the shared data is available.

You can access the shared data in any template by using the key names.

```html
<h1>{{ siteName }}</h1>
```

Returning a function from a `SharedContextProviderFunc` is a good practice because it enables lazy evaluation of the shared data,
reducing the overhead associated with the shared context.

```go
v := viewkit.New()
v.SharedContextProviders = map[string]viewkit.SharedContextProviderFunc{
	"siteName": func(c echo.Context) (any, error) {
		return func() string {
			return "Echo ViewKit Example"
		}, nil
	},
}
```

```html
{# You can call the function like this. #}
<h1>{{ siteName() }}</h1>

{# Or you can still access like a value. #}
{# Because you can omit the brackets when calling a function without arguments. #}
<h1>{{ siteName }}</h1>
```

### Standard shared context providers

Echo ViewKit has several standard shared context providers and automatically sets them up by default.
You can use the following functions in any template without any additional configuration.

Here are the functions provided by standard shared context providers:

- [`is_debug`](#is-debug)
- [`url_path`](#url-path)
- [`url_query`](#url-query)
- [`url_path_query`](#url-path-query)
- [`json_marshal`](#json-marshal)

#### is_debug

A function that returns `true` if you set Echo ViewKit's [`Debug` property](#debug-mode) to `true`.

```html
{% if is_debug %}
  <p>This is a debug mode.</p>
{% endif %}
```

#### url_path

A function that returns the path of the current URL.

```html
<p>{{ url_path }}</p> {# => /path/to/current #}
```

#### url_query

A function that returns the query string of the current URL.

```html
<p>{{ url_query }}</p> {# => key1=value1&key2=value2&key3=value3 #}
<p>{{ url_query("key1") }}</p> {# => value1 #}
```

#### url_path_query

A function that returns the path and query string of the current URL.

```html
<p>{{ url_path_query }}</p> {# => /path/to/current?key1=value1&key2=value2&key3=value3 #}
<p>{{ url_path_query("key1") }}</p> {# => /path/to/current?key1=value1 #}
<p>{{ url_path_query("key1", "key2") }}</p> {# => /path/to/current?key1=value1&key2=value2 #}
```

#### json_marshal

A function that returns the JSON string of the passed value.

```html
<div data-json="{{ json_marshal(value) }}"></div>
```

### Disabling standard shared context providers

If you don't want to automatically apply the standard shared context providers,
you can disable them by setting the `DisableStandardSharedContextProviders` property to `true`.

```go
v := viewkit.New()
v.DisableStandardSharedContextProviders = true
```

### Share passed data

You can also share the data passed to the `Render` method by registering key names to the `SharedContextKeys` property.

```go
v := viewkit.New()
v.SharedContextKeys = []string{"message"}
```

If you pass the data with the key name `message`, it can be accessed in all templates.

```go
c.Render(http.StatusOK, "hello", map[string]any{
	"message": "Hello World",
})
```

## View handlers

Sometimes, you don't need a dedicated handler function for a view template.
In that case, you can use the `ViewHandler` to render the template directly.

```go
e.GET("/", viewkit.ViewHandler("pages/index"))
```

If you want to pass data to the template, you can use the `ViewHandlerWithData` function.

```go
e.GET("/", viewkit.ViewHandlerWithData("pages/index", map[string]any{
  "message": "Hello World",
}))
```

## Rendering fragments

You can render a fragment of a template using the `fragment` template tag.
This is especially useful when working with frontend frameworks like [htmx](https://htmx.org/), as these types of frameworks load portions of the page dynamically.

In your template, you can define a fragment like this:

```html
{% fragment "user-list" %}
  <ul>
    {% for user in users %}
      <li>{{ user.Name }}</li>
    {% endfor %}
  </ul>
{% endfragment %}
```

Then, you can render only the fragment by specifying its name after the template name, using the `#` separator, like this:

```go
c.Render(http.StatusOK, "index#user-list", map[string]any{
  "users": users,
})
```
