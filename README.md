# Echo ViewKit

A comprehensive view package for [Echo](https://github.com/labstack/echo).

## Overview

Echo is a high performance, extensible, minimalist Go web framework.
It is an excellent choice for building web API servers.
Echo also supports rendering HTML templates.
The official documentation explains [how to implement template rendering](https://echo.labstack.com/docs/templates) using the standard `html/template` package.
However, this way is insufficient for building most real-world web applications, because the `html/template` package lacks many features commonly needed for modern web development.

Echo ViewKit is created to solve this problem. It provides the following features:

- **Powerful templating engine**:
[Pongo2](https://github.com/flosch/pongo2), the core of Echo ViewKit, is a Django-like syntax template engine for Go.
Furthermore, we use a forked version of Pongo2, extensively customizing it to meet the demands of modern web development and seamless Echo integration.
- **Component-based architecture**:
Our templating system is also inspired by [Laravel Blade](https://laravel.com/docs/11.x/blade), another templating engine bundled with Laravel PHP framework.
Echo ViewKit provides a component-based architecture like [Laravel Blade Components](https://laravel.com/docs/11.x/blade#components). It enhances code organization and maintainability.
- **Vite integration**:
Front-end build tools are essential for modern web development.
Echo ViewKit provides a [Vite](https://vite.dev/) integration that allows you to use Vite for building front-end assets.

## Website and Documentation

[https://echo-viewkit.kohkimakimoto.dev](https://echo-viewkit.kohkimakimoto.dev)

## Author

Kohki Makimoto <kohki.makimoto@gmail.com>

## License

The MIT License (MIT)
