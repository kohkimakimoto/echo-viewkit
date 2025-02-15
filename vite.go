package viewkit

import (
	"encoding/json"
	"fmt"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/labstack/echo/v4"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"
)

// Vite integration
// see also: https://vite.dev/guide/backend-integration.html

func ViteFunctionProvider(v *ViewKit) SharedContextProviderFunc {
	return func(c echo.Context) (any, error) {
		return func(entryPoints ...string) (*pongo2.Value, error) {
			if v.ViteDevMode {
				tags := []string{
					fmt.Sprintf(`<script type="module" src="%s/@vite/client"></script>`, strings.TrimSuffix(v.ViteDevServerURL, "/")),
				}
				for _, entryPoint := range entryPoints {
					tags = append(tags, genViteAssetTag(fmt.Sprintf("%s/%s", strings.TrimSuffix(v.ViteDevServerURL, "/"), strings.TrimPrefix(entryPoint, "/"))))
				}
				return pongo2.AsSafeValue(strings.Join(tags, "")), nil
			}

			if v.ViteManifest == nil {
				return nil, fmt.Errorf("the Vite manifest is not loaded")
			}

			tags := []string{}
			for _, entryPoint := range entryPoints {
				chunk, ok := v.ViteManifest[entryPoint]
				if !ok {
					return nil, fmt.Errorf("the Vite manifest does not have the entrypoint: %s", entryPoint)
				}

				if chunk, ok := chunk.(map[string]any); ok {
					file := chunk["file"].(string)
					tags = append(tags, genViteAssetTag(path.Join(v.ViteBasePath, file)))
					if cssList, ok := chunk["css"].([]any); ok {
						for _, cssV := range cssList {
							cssFile, ok := cssV.(string)
							if !ok {
								return nil, fmt.Errorf("the Vite manifest has an invalid css file: %v", cssV)
							}
							tags = append(tags, genViteAssetTag(path.Join(v.ViteBasePath, cssFile)))
						}
					}
				}
			}
			return pongo2.AsSafeValue(strings.Join(tags, "")), nil
		}, nil
	}
}

func ViteReactRefreshFunctionProvider(v *ViewKit) SharedContextProviderFunc {
	return func(c echo.Context) (any, error) {
		return func() *pongo2.Value {
			if v.ViteDevMode {
				return pongo2.AsSafeValue(fmt.Sprintf(`<script type="module">import RefreshRuntime from '%s/@react-refresh';RefreshRuntime.injectIntoGlobalHook(window);window.$RefreshReg$ = () => {};window.$RefreshSig$ = () => (type) => type;window.__vite_plugin_react_preamble_installed__ = true;</script>`, v.ViteDevServerURL))
			}
			return pongo2.AsSafeValue("")
		}, nil
	}
}

func genViteAssetTag(url string) string {
	if isCssPath(url) {
		return fmt.Sprintf(`<link rel="stylesheet" href="%s" />`, url)
	} else {
		return fmt.Sprintf(`<script type="module" src="%s"></script>`, url)
	}
}

var cssRe = regexp.MustCompile(`\.(css|less|sass|scss|styl|stylus|pcss|postcss)$`)

func isCssPath(url string) bool {
	return cssRe.MatchString(url)
}

type ViteManifest map[string]any

func ParseViteManifest(data []byte) (ViteManifest, error) {
	var manifest ViteManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the Vite manifest: %w", err)
	}
	return manifest, nil
}

func MustParseViteManifest(data []byte) ViteManifest {
	manifest, err := ParseViteManifest(data)
	if err != nil {
		panic(err)
	}
	return manifest
}

func ParseViteManifestFile(name string) (ViteManifest, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read the Vite manifest from: %s: %w", name, err)
	}
	return ParseViteManifest(data)
}

func MustParseViteManifestFile(name string) ViteManifest {
	manifest, err := ParseViteManifestFile(name)
	if err != nil {
		panic(err)
	}
	return manifest
}

func ParseViteManifestFS(f fs.FS, name string) (ViteManifest, error) {
	data, err := fs.ReadFile(f, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read the Vite manifest from: %s: %w", name, err)
	}
	return ParseViteManifest(data)
}

func MustParseViteManifestFS(f fs.FS, name string) ViteManifest {
	manifest, err := ParseViteManifestFS(f, name)
	if err != nil {
		panic(err)
	}
	return manifest
}
