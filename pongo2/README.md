# Pongo2

This `pongo2` package is a fork of the [pongo2](https://github.com/flosch/pongo2).

## Why forked?

- I want to add some features that are not available in the original pongo2 for my own use.
- Original pongo2 is not maintained for a long time.

## Key changes from the original pongo2

This fork of the source code is based on the pongo2 commit [c84aecb](https://github.com/flosch/pongo2/commit/c84aecb5fa79a9c0feec284a7bf4f0536c6a6e99),
and I have made the following changes:

- Make tags and filters private for templateSet. It is related to the PR [https://github.com/flosch/pongo2/pull/335](https://github.com/flosch/pongo2/pull/335).
- Add `OmitExtensionLoader` to omit the extension of the template name.
- Add `PreProcessLoader` to implement the pre-process of the template content before parsing.
- Add `component` tag. It is inspired by the [django-component](https://github.com/EmilStenstrom/django-components).

... and more.

TODO: Add more details.

## Original pongo2 README

You can read the original pongo2 README as [README.pongo2.md](README.pongo2.md).
