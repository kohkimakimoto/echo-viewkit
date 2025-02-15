---
title: Components
---

# Components

Component-based architecture is one of the most common patterns in modern web development.
You can see it in frameworks like React, Vue and also in backend frameworks like [Laravel](https://laravel.com/docs/11.x/blade#components), [Django](https://github.com/EmilStenstrom/django-components), and others.

Echo ViewKit supports similar component-based architecture in Pongo2 templates.

## Introduction

To use components in your templates, you need to register them first.
For example, create a new component named `greeting` in your Go code:

```go
import "github.com/kohkimakimoto/echo-viewkit/pongo2"

var Greeting = &pongo2.Component{
	// The name of the component.
	Name:         "greeting",
	// The path to the template file rendered by the component.
	TemplateFile: "components/greeting",
	// Define the properties that can be passed to the template.
	Props:        []string{"color", "message"},
	// The setup function is called before rendering the component.
	Setup: func(ctx *pongo2.ComponentExecutionContext) error {
		// You can set data to the component context.
		ctx.Default("color", "blue")
		ctx.Default("message", "Hello, World!")
		return nil
	},
}
```

And then, create the corresponding template file `views/components/greeting.html`:

```html
<div style="color: {{ color }};">{{ message }}</div>
```

Finally, register the component in your Echo application:

```go
v := viewkit.New()
v.BaseDir = "views"
v.Components = []*pongo2.Component{
	Greeting,
}
e.Renderer = v.MustRenderer()
```

Now you can use the `greeting` component in your templates like the following:

```html
<x-greeting/>
```

The above code will render the following HTML:

```html
<div style="color: blue;">Hello, World!</div>
```

## Rendering components

you can render components by using the `<x-component-name>` tag in your templates.

For example, if you have a component named `alert`, you can render it like this:

```html
<x-alert/>
<!-- or -->
<x-alert></x-alert>
```

## Passing data to components

To make your components reusable, you can pass data (properties) to them.
Components require explicit props declaration by setting the `Props` field in the component definition.

```go
var Alert = &pongo2.Component{
	Props: []string{"text", "color"},
	// ...
}
````

And then, you can pass data to components by using attributes.
If you want to pass expressions and variables to components, you can use the `:` prefix.

```html
<x-alert text="Error message here!" color="red"/>
<!-- use variable -->
<x-alert :text="text" :color="color"/>
```

These attributes can be accessed in the component template like this:

```html
<div style="color: {{ color }};">{{ text }}</div>
```

If you pass data with kebab-case, you can access it with camelCase.
For example, if you pass `text-message`, you can access it with `textMessage`.

```html
<!-- <x-alert text-message="text message here!"/> -->
<div>{{ textMessage }}</div>
```

## Component attributes

Sometimes, you may need to pass arbitrary attributes that are not defined in the `Props` field.
For example, you might want to pass a `class` attribute that is not essential for the component's functionality.
However, you still want to customize the component's appearance by specifying a `class` attribute.

```go
var Alert = &pongo2.Component{
	Props: []string{"text"},
	// ...
}
````

```html
<x-alert :text="text" class="mt-4"/>
```

To render these attributes, you can use the `attributes` variable in the component template.

```html
<div {{ attributes }}>{{ text }}</div>
```

This will render the following HTML:

```html
<div class="mt-4">...</div>
```

### Default

In your component template, you can set defaults that are merged with passed attributes.
The passed attributes override the default except `class`. For `class`, the defaults are prepended:

```html
<div {{ attributes.Default("class", "bg-white") }}></div>
<!-- Render as: -->
<!-- <div class="bg-white mt-4"></div> -->

<!-- The method can be chained -->
<div {{ attributes.Default("class", "bg-white").Default('id', 'foo') }}></div>
<!-- Render as: -->
<!-- <div class="bg-white mt-4" id="foo"></div> -->
```

### Only

Only method extracts specific attributes:

```html
<!-- Render class attribute only -->
<div {{ attributes.Only("class") }}></div>
<!-- Render as: -->
<!-- <div class="mt-4"></div> -->

<!-- Render id and class attributes only -->
<div {{ attributes.Only("id", "class") }}></div>
<!-- Render as: -->
<!-- <div id="foo" class="mt-4"></div> -->
```

### Without

Without method excludes specific attributes:

```html
<!-- Render all attributes except class -->
<div {{ attributes.Without("class") }}></div>
<!-- Render as: -->
<!-- <div id="foo"></div> -->

<!-- Render all attributes except id and class -->
<div {{ attributes.Without("id", "class") }}></div>
<!-- Render as: -->
<!-- <div></div> -->
```

### Get

Get method gets the value of the specified attribute:

```html
<!-- Get the value of the class attribute: -->
{{ attributes.Get("class") }}
<!-- Render as: -->
<!-- mt-4 -->
```

### Has

Has method checks if the specified attribute exists:

```html
{% if attributes.Has("class") %}
  Passed class attribute: {{ attributes.Get("class") }}
{% endif %}
```

## Slots

Slots are a way to pass content to a component.
The `slot` variable in the component template contains the content passed to the component.

```html
<!-- views/components/alert.html -->
<div class="alert">
  {{ slot }}
</div>
```

You can pass content to the component like this:

```html
<!-- views/index.html -->
<x-alert>
  <p>Alert message here!</p>
</x-alert>
```

This will render the following HTML:

```html
<div class="alert">
  <p>Alert message here!</p>
</div>
```

### Named slots

You can also pass named slots to the component using the `x-slot` tag:

```html
<!-- views/components/modal.html -->
<div>
  <div class="header">
    {{ header }}
  </div>
  <div class="body">
    {{ body }}
  </div>
</div>
```

This component can be used like this:

```html
<!-- views/index.html -->
<x-modal>
  <x-slot name="header">
    <h2>Modal header</h2>
  </x-slot>
  <x-slot name="body">
    <p>Modal body</p>
  </x-slot>
</x-modal>
```

This will render the following HTML:

```html
<div>
  <div class="header">
    <h2>Modal header</h2>
  </div>
  <div class="body">
    <p>Modal body</p>
  </div>
</div>
```

### Scoped slots

Scoped slots allow you to access data from the component within your slot.
For example, the component provides data such as `message` and `type` to the slot:

```go
var Alert = &pongo2.Component{
	Name: "alert",
	Props: []string{"message", "type"},
	TemplateFile: "components/alert",
	Setup: func(ctx *pongo2.ComponentExecutionContext) error {
		ctx.Default("message", "Hello, World!")
		ctx.Default("type", "warning")
		return nil
  },
}
```

```html
<!-- views/components/alert.html -->
<div class="alert alert-{{ type }}">
  {{ slot }}
</div>
```

You can use this component and access the data using the `slot-data` attribute like the following:

```html
<!-- views/index.html -->
<x-alert slot-data="{message, type}">
  <p>{{ type }}: {{ message }}</p>
</x-alert>

<!-- You can assign different names to the slot data -->
<x-alert slot-data="{message: msg, type: t}">
  <p>{{ t }}: {{ msg }}</p>
</x-alert>
```

## Setup function

The `Setup` function of the component is called before rendering the component.
For example, if you use the component like this:

```html
<x-alert message="This is an alert message." class="text-red-500" />
```

You can access the passed props and attributes in the `Setup` function:

```go
var Alert = &pongo2.Component{
	Props: []string{"message", "type"},
	Setup: func (ctx *pongo2.ComponentExecutionContext) error {
		// Get "This is an alert message." from the message prop
		message := ctx.Get("message")
		// ...
		return nil
	},
}
```

The `Setup` function is a good place to initialize or configure the data.
You can use helper methods for working with component data like the following:

```go
// Set the default value for the message prop
ctx.Default("message", "The default message")

// Override the message prop
ctx.Set("message", "The new message")

// Delete the message prop
ctx.Delete("message")

// Get component attributes
attributes := ctx.Attributes()
```

If you prefer using a struct type, you can define a struct for component props and use the following methods:

```go
// Define the AlertProps struct
type AlertProps struct {
	Message string `pongo2:"message"`
	Type    string `pongo2:"type"`
}

// In the Setup function

// Set default values for the props
if err := ctx.Defaults(&AlertProps{
	Message: "This is an alert message.",
	Type: "info",
}); err != nil {
	return err
}

// Get the props
props := &AlertProps{}
if err := ctx.Bind(props); err != nil {
	return err
}

// Override the props
if err := ctx.Update(&AlertProps{
	Message: "The new message",
	Type: "warning",
}); err != nil {
	return err
}
```

You can also access the Echo context by `ctx.EchoContext` property.

```go
c := ctx.EchoContext
c.Logger().Info("This is a log message")
```

## Inline components

If you have a simple component that doesn't require a separate template file, you can define it inline.

```go
var InlineAlert = &pongo2.InlineComponent{
	Name:  "alert",
	Props: []string{"message", "type"},
	TemplateString: `
		<div class="alert alert-{{ type }}">
			{{ message }}
		</div>
	`,
}
```

Yon have to register your inline components like this:

```go
v := viewkit.New()
v.BaseDir = "views"
v.InlineComponents = []*pongo2.InlineComponent{
	InlineAlert,
}
```

## Anonymous components

You can define a component using a single template file without a separate Go source file.

```html
<!-- views/components/primary-button.html -->
<button {{ attributes.Default("class", "primary") }}>
  {{ slot }}
</button>
```

You have to register your anonymous components like this:

```go
v := viewkit.New()
v.BaseDir = "views"
v.AnonymousComponents = []*pongo2.AnonymousComponent{
	{Name: "primary-button", TemplateFile: "components/primary-button"},
}
// As a result, you can use the component like this:
// <x-primary-button>Click me</x-primary-button>
```

You can also register anonymous components from specific directories:

```go
v := viewkit.New()
v.BaseDir = "views"
v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
	{Dir: "components"},
	// You can register multiple directories.
	// A prefix is an optional setting that helps prevent naming conflicts
	// by adding a specified string to the beginning of the component name.
	{Dir: "ui-components", Prefix: "ui."},
}
```

The above example will register all template files in the `views/components` directory as anonymous components.
The component name is generated from the file path by converting the path separator to a dot.

For examples:

- filepath:`views/components/primary-button.html` => component name:`primary-button`, tag name:`<x-primary-button>`
- filepath:`views/components/button/primary.html` => component name:`button.primary`, tag name:`<x-button.primary>`

With the `Prefix: "ui."`:

- filepath:`views/ui-components/primary-button.html` => component name:`ui.primary-button`, tag name:`<x-ui.primary-button>`
- filepath:`views/ui-components/button/primary.html` => component name:`ui.button.primary`, tag name:`<x-ui.button.primary>`

### Props and attributes

Since anonymous components do not have any associated Go code, you need to define the props and attributes in the template file.
To do this, you can use the `props` template tag:

```html
{% props message, type %}

<div {{ attributes.Default("class", "alert alert-" + type) }}>
  {{ message }}
</div>
```

The above example defines the `message` and `type` as props and others as attributes.
So you can use the component like this:

```html
<x-alert message="This is an alert message." type="info" class="text-red-500" />
```

The `props` template tag can be set with default values:

```html
{% props message="This is an alert message.", type="info" %}
```
