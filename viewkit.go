package viewkit

import (
	"fmt"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/kohkimakimoto/echo-viewkit/subprocess"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type ViewKit struct {
	// Debug enables debug mode.
	Debug bool

	// Templates

	// FS is a file system to load templates.
	FS fs.FS
	// FSBaseDir is a root directory to load templates from the FS.
	// If you want to load templates from a subdirectory of the FS instead of the root directory, set this option.
	// This option is used with the FS property.
	FSBaseDir string
	// BaseDir is a root directory to load templates directly from the local file system.
	// It is used when FS is not set.
	BaseDir string
	// DefaultTemplateFileExtension is a file extension that can be omitted when loading templates.
	// This option is used when loading templates without specifying a file extension, such as “index”.
	// The default value is “.html”.
	DefaultTemplateFileExtension string
	// PreProcessors is a list of PreProcessor.
	PreProcessors []pongo2.PreProcessor
	// Filters is a map of filters to be registered.
	Filters map[string]pongo2.FilterFunction
	// Tags is a map of tags to be registered.
	Tags map[string]pongo2.TagParser
	// Component config

	// DisableComponentHTMLTag disables the HTML syntax extension for components.
	DisableComponentHTMLTag bool
	// ComponentHTMLTagPrefix is a prefix for the component tag.
	// The default value is “x-”.
	ComponentHTMLTagPrefix string

	// Components

	// AnonymousComponentsDirectories is a list of components directories.
	AnonymousComponentsDirectories []*pongo2.AnonymousComponentsDirectory
	// AnonymousComponents is a list of template file components.
	AnonymousComponents []*pongo2.AnonymousComponent
	// HeadlessComponents is a list of headless components.
	HeadlessComponents []*pongo2.HeadlessComponent
	// InlineComponents is a list of inline components.
	InlineComponents []*pongo2.InlineComponent
	// Components is a list of components.
	Components []*pongo2.Component
	// Shared context

	// SharedContextProviders is a map of shared context providers.
	// The map keys are accessible from all templates.
	SharedContextProviders map[string]SharedContextProviderFunc
	// DisableStandardSharedContextProviders disables loading standard context providers that are provided by the package.
	DisableStandardSharedContextProviders bool
	// SharedContextKeys includes additional keys that are accessible from all templates.
	SharedContextKeys []string

	// Vite integration
	// see also: https://vite.dev/guide/backend-integration.html

	// If you want to use Vite, set this to true.
	Vite bool
	// ViteDevMode enables Vite development mode.
	// It means that the server uses the Vite development server to resolve the assets.
	ViteDevMode bool
	// ViteDevServerURL is a base URL of the Vite development server.
	// Default is "http://localhost:5173",
	ViteDevServerURL string
	// ViteDevServerCommand is a command to start the Vite dev server.
	// The default value is []string{"npx", "vite", "--clearScreen=false"}.
	ViteDevServerCommand []string
	// ViteDevServerStdout is a writer for the Vite dev server stdout.
	// The default value is os.Stdout.
	ViteDevServerStdout io.Writer
	// ViteDevServerStderr is a writer for the Vite dev server stderr.
	// The default value is os.Stderr.
	ViteDevServerStderr io.Writer
	// ViteDevServerLogPrefix is a prefix for the Vite dev server log.
	// The default value is "[vite] ".
	ViteDevServerLogPrefix string
	// ViteManifest is a Vite manifest.
	// This is needed to resolve the asset paths in production environment.
	ViteManifest ViteManifest
	// ViteBasePath is a base path for the built assets.
	ViteBasePath string

	// The renderer instance
	renderer *Renderer
}

func New() *ViewKit {
	return &ViewKit{
		Debug:                                 false,
		FS:                                    nil,
		FSBaseDir:                             "",
		BaseDir:                               "",
		DefaultTemplateFileExtension:          ".html",
		PreProcessors:                         []pongo2.PreProcessor{},
		Filters:                               map[string]pongo2.FilterFunction{},
		Tags:                                  map[string]pongo2.TagParser{},
		DisableComponentHTMLTag:               false,
		ComponentHTMLTagPrefix:                "x-",
		Components:                            []*pongo2.Component{},
		AnonymousComponentsDirectories:        []*pongo2.AnonymousComponentsDirectory{},
		InlineComponents:                      []*pongo2.InlineComponent{},
		HeadlessComponents:                    []*pongo2.HeadlessComponent{},
		SharedContextKeys:                     []string{},
		DisableStandardSharedContextProviders: false,
		SharedContextProviders:                map[string]SharedContextProviderFunc{},
		Vite:                                  false,
		ViteDevMode:                           false,
		ViteDevServerURL:                      "http://localhost:5173",
		ViteDevServerCommand:                  []string{"npx", "vite", "--clearScreen=false"},
		ViteDevServerStdout:                   os.Stdout,
		ViteDevServerStderr:                   os.Stderr,
		ViteDevServerLogPrefix:                "[echo-viewkit:vite] ",
		ViteManifest:                          nil,
		ViteBasePath:                          "",
	}
}

func (v *ViewKit) Renderer() (*Renderer, error) {
	if v.renderer != nil {
		return v.renderer, nil
	}

	var loader pongo2.TemplateLoader
	var templateSetFS fs.FS

	if v.FS != nil {
		// use FS loader
		var tFs fs.FS
		if v.FSBaseDir != "" {
			subFs, err := fs.Sub(v.FS, filepath.ToSlash(filepath.Clean(v.FSBaseDir)))
			if err != nil {
				return nil, fmt.Errorf("failed to create sub fs: %w", err)
			}
			tFs = subFs
		} else {
			tFs = v.FS
		}
		loader = pongo2.NewFSLoader(tFs)
		templateSetFS = tFs
	} else if v.BaseDir != "" {
		// use local file system loader
		l, err := pongo2.NewLocalFileSystemLoader(v.BaseDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create local file system loader: %w", err)
		}
		loader = l
		templateSetFS = os.DirFS(v.BaseDir)
	} else {
		return nil, fmt.Errorf("FS or BaseDir is required")
	}

	if v.DefaultTemplateFileExtension != "" {
		// use omit extension loader to load templates without the default file extension.
		loader = pongo2.NewOmitExtensionLoader(loader, v.DefaultTemplateFileExtension)
	}

	// pre processors
	preProcessors := v.PreProcessors
	if v.ComponentHTMLTagPrefix != "" && !v.DisableComponentHTMLTag {
		preProcessors = append(preProcessors, pongo2.ComponentHTMLTagPreProcessor(pongo2.ComponentHTMLTagPreProcessorConfig{
			TagPrefix: v.ComponentHTMLTagPrefix,
		}))
	}

	if len(preProcessors) > 0 {
		loader = pongo2.NewPreProcessLoader(loader, preProcessors...)
	}

	// template set
	ts := pongo2.NewSet("renderer", loader)
	ts.Debug = v.Debug

	// configuration for components
	ts.ComponentSet.TemplateSetFS = templateSetFS
	ts.ComponentSet.DefaultTemplateFileExtension = v.DefaultTemplateFileExtension

	// register filters
	for name, filter := range v.Filters {
		if err := ts.RegisterFilter(name, filter); err != nil {
			return nil, fmt.Errorf("failed to register filter %s: %w", name, err)
		}
	}

	// register tags
	for name, tag := range v.Tags {
		if err := ts.RegisterTag(name, tag); err != nil {
			return nil, fmt.Errorf("failed to register tag %s: %w", name, err)
		}
	}

	// register components
	for _, dir := range v.AnonymousComponentsDirectories {
		if err := ts.ComponentSet.RegisterTemplateFileComponentsDirectory(dir); err != nil {
			return nil, err
		}
	}
	for _, comp := range v.AnonymousComponents {
		ts.ComponentSet.RegisterAnonymousComponent(comp)
	}
	for _, comp := range v.HeadlessComponents {
		ts.ComponentSet.RegisterHeadlessComponent(comp)
	}
	for _, comp := range v.InlineComponents {
		ts.ComponentSet.RegisterInlineComponent(comp)
	}
	for _, comp := range v.Components {
		ts.ComponentSet.RegisterComponent(comp)
	}

	// Shared context configuration
	sharedContextKeys := []string{}
	sharedContextProviders := map[string]SharedContextProviderFunc{}

	if !v.DisableStandardSharedContextProviders {
		// Apply standard context providers
		sharedContextProviders["is_debug"] = IsDebugFunctionProvider(v.Debug)
		sharedContextProviders["url_path"] = URLPathFunctionProvider()
		sharedContextProviders["url_query"] = URLQueryFunctionProvider()
		sharedContextProviders["url_path_query"] = URLPathQueryFunctionProvider()
		sharedContextProviders["json_marshal"] = JsonMarshalFunctionProvider()
	}

	if v.Vite {
		sharedContextProviders["vite"] = ViteFunctionProvider(v)
		sharedContextProviders["vite_react_refresh"] = ViteReactRefreshFunctionProvider(v)
	}

	// merge user defined shared context providers
	sharedContextProviders = mergeSharedContextProviders(sharedContextProviders, v.SharedContextProviders)

	// append keys that provided by the providers
	for k := range sharedContextProviders {
		sharedContextKeys = append(sharedContextKeys, k)
	}

	// append user defined shared context keys and remove duplicated keys
	ts.SharedContextKeys = uniqueStrings(append(sharedContextKeys, v.SharedContextKeys...))

	v.renderer = &Renderer{
		templateSet: ts,
		providers:   sharedContextProviders,
	}

	return v.renderer, nil
}

func (v *ViewKit) MustRenderer() *Renderer {
	r, err := v.Renderer()
	if err != nil {
		panic(err)
	}
	return r
}

func (v *ViewKit) StartViteDevServer() error {
	if !v.Vite {
		return fmt.Errorf("vite integration is disabled")
	}

	return subprocess.Run(&subprocess.Subprocess{
		Command:   v.ViteDevServerCommand[0],
		Args:      v.ViteDevServerCommand[1:],
		Stdout:    v.ViteDevServerStdout,
		Stderr:    v.ViteDevServerStderr,
		LogPrefix: v.ViteDevServerLogPrefix,
	})
}

func uniqueStrings(s []string) []string {
	m := make(map[string]struct{})
	var r []string
	for _, v := range s {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
			r = append(r, v)
		}
	}
	return r
}

func mergeSharedContextProviders(providers ...map[string]SharedContextProviderFunc) map[string]SharedContextProviderFunc {
	m := map[string]SharedContextProviderFunc{}
	for _, p := range providers {
		for k, v := range p {
			m[k] = v
		}
	}
	return m
}
