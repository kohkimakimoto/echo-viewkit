package pongo2

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// TemplateLoader allows to implement a virtual file system.
type TemplateLoader interface {
	// Abs calculates the path to a given template. Whenever a path must be resolved
	// due to an import from another template, the base equals the parent template's path.
	Abs(base, name string) string

	// Get returns an io.Reader where the template's content can be read from.
	Get(path string) (io.Reader, error)
}

// TemplateSet allows you to create your own group of templates with their own
// global context (which is shared among all members of the set) and their own
// configuration.
// It's useful for a separation of different kind of templates
// (e. g. web templates vs. mail templates).
type TemplateSet struct {
	name    string
	loaders []TemplateLoader

	// Globals will be provided to all templates created within this template set
	Globals Context

	// If debug is true (default false), ExecutionContext.Logf() will work and output
	// to STDOUT. Furthermore, FromCache() won't cache the templates.
	// Make sure to synchronize the access to it in case you're changing this
	// variable during program execution (and template compilation/execution).
	Debug bool

	// Options allow you to change the behavior of template-engine.
	// You can change the options before calling the Execute method.
	Options *Options

	// Sandbox features
	// - Disallow access to specific tags and/or filters (using BanTag() and BanFilter())
	//
	// For efficiency reasons you can ban tags/filters only *before* you have
	// added your first template to the set (restrictions are statically checked).
	// After you added one, it's not possible anymore (for your personal security).
	tags          map[string]*tag
	filters       map[string]FilterFunction
	bannedTags    map[string]bool
	bannedFilters map[string]bool

	// Template cache (for FromCache())
	templateCache      map[string]*Template
	templateCacheMutex sync.RWMutex

	// SharedContextKeys is a list of shared context keys.
	// These keys can be accessible from all templates.
	SharedContextKeys []string

	// components
	ComponentSet *componentSet
}

// newSet only be used to create sets without default tags and filters
func newSet(name string, loaders ...TemplateLoader) *TemplateSet {
	if len(loaders) == 0 {
		panic(fmt.Errorf("at least one template loader must be specified"))
	}

	return &TemplateSet{
		name:              name,
		loaders:           loaders,
		Globals:           make(Context),
		tags:              make(map[string]*tag),
		filters:           make(map[string]FilterFunction),
		bannedTags:        make(map[string]bool),
		bannedFilters:     make(map[string]bool),
		templateCache:     make(map[string]*Template),
		Options:           newOptions(),
		SharedContextKeys: []string{},
		ComponentSet:      newComponentSet(),
	}
}

// NewSet can be used to create sets with different kind of templates
// (e. g. web from mail templates), with different globals or
// other configurations.
func NewSet(name string, loaders ...TemplateLoader) *TemplateSet {
	set := newSet(name, loaders...)

	for name, tag := range DefaultSet.tags {
		set.tags[name] = tag
	}
	for name, filter := range DefaultSet.filters {
		set.filters[name] = filter
	}
	return set
}

func (set *TemplateSet) AddLoader(loaders ...TemplateLoader) {
	set.loaders = append(set.loaders, loaders...)
}

func (set *TemplateSet) resolveFilename(tpl *Template, path string) string {
	return set.resolveFilenameForLoader(set.loaders[0], tpl, path)
}

func (set *TemplateSet) resolveFilenameForLoader(loader TemplateLoader, tpl *Template, path string) string {
	name := ""
	if tpl != nil && tpl.isTplString {
		return path
	}
	if tpl != nil {
		name = tpl.name
	}

	return loader.Abs(name, path)
}

// BanTag bans a specific tag for this template set. See more in the documentation for TemplateSet.
// This method must be called before you've added your first template to the set.
func (set *TemplateSet) BanTag(name string) error {
	_, has := set.tags[name]
	if !has {
		return fmt.Errorf("tag '%s' not found", name)
	}
	_, has = set.bannedTags[name]
	if has {
		return fmt.Errorf("tag '%s' is already banned", name)
	}
	set.bannedTags[name] = true

	return nil
}

// BanFilter bans a specific filter for this template set. See more in the documentation for TemplateSet.
// This method must be called before you've added your first template to the set.
func (set *TemplateSet) BanFilter(name string) error {
	_, has := set.filters[name]
	if !has {
		return fmt.Errorf("filter '%s' not found", name)
	}
	_, has = set.bannedFilters[name]
	if has {
		return fmt.Errorf("filter '%s' is already banned", name)
	}
	set.bannedFilters[name] = true

	return nil
}

func (set *TemplateSet) resolveTemplate(tpl *Template, path string) (name string, loader TemplateLoader, fd io.Reader, err error) {
	// iterate over loaders until we appear to have a valid template
	for _, loader = range set.loaders {
		name = set.resolveFilenameForLoader(loader, tpl, path)
		fd, err = loader.Get(name)
		if err == nil {
			return
		}
	}

	return path, nil, nil, fmt.Errorf("unable to resolve template")
}

// CleanCache cleans the template cache. If filenames is not empty,
// it will remove the template caches of those filenames.
// Or it will empty the whole template cache. It is thread-safe.
func (set *TemplateSet) CleanCache(filenames ...string) {
	set.templateCacheMutex.Lock()
	defer set.templateCacheMutex.Unlock()

	if len(filenames) == 0 {
		set.templateCache = make(map[string]*Template, len(set.templateCache))
	}

	for _, filename := range filenames {
		delete(set.templateCache, set.resolveFilename(nil, filename))
	}
}

// FromCache is a convenient method to cache templates. It is thread-safe
// and will only compile the template associated with a filename once.
// If TemplateSet.Debug is true (for example during development phase),
// FromCache() will not cache the template and instead recompile it on any
// call (to make changes to a template live instantaneously).
func (set *TemplateSet) FromCache(filename string) (*Template, error) {
	if set.Debug {
		// Recompile on any request
		return set.FromFile(filename)
	}
	// Cache the template
	cleanedFilename := set.resolveFilename(nil, filename)

	set.templateCacheMutex.RLock()
	tpl, has := set.templateCache[cleanedFilename]
	set.templateCacheMutex.RUnlock()

	// Cache hit
	if has {
		return tpl, nil
	}

	// Cache miss and acquire the lock
	set.templateCacheMutex.Lock()
	defer set.templateCacheMutex.Unlock()

	// Check again if the template has been added to the cache by another goroutine.
	tpl, has = set.templateCache[cleanedFilename]
	if has {
		return tpl, nil
	}

	tpl, err := set.FromFile(cleanedFilename)
	if err != nil {
		return nil, err
	}
	set.templateCache[cleanedFilename] = tpl
	return tpl, nil
}

// FromString loads a template from string and returns a Template instance.
func (set *TemplateSet) FromString(tpl string) (*Template, error) {
	return newTemplateString(set, []byte(tpl))
}

// FromBytes loads a template from bytes and returns a Template instance.
func (set *TemplateSet) FromBytes(tpl []byte) (*Template, error) {
	return newTemplateString(set, tpl)
}

// FromFile loads a template from a filename and returns a Template instance.
func (set *TemplateSet) FromFile(filename string) (*Template, error) {
	_, _, fd, err := set.resolveTemplate(nil, filename)
	if err != nil {
		return nil, &Error{
			Filename:  filename,
			Sender:    "fromfile",
			OrigError: err,
		}
	}
	buf, err := io.ReadAll(fd)
	if err != nil {
		return nil, &Error{
			Filename:  filename,
			Sender:    "fromfile",
			OrigError: err,
		}
	}

	return newTemplate(set, filename, false, buf)
}

// RenderTemplateString is a shortcut and renders a template string directly.
func (set *TemplateSet) RenderTemplateString(s string, ctx Context) (string, error) {
	tpl := Must(set.FromString(s))
	result, err := tpl.Execute(ctx)
	if err != nil {
		return "", err
	}
	return result, nil
}

// RenderTemplateBytes is a shortcut and renders template bytes directly.
func (set *TemplateSet) RenderTemplateBytes(b []byte, ctx Context) (string, error) {
	tpl := Must(set.FromBytes(b))
	result, err := tpl.Execute(ctx)
	if err != nil {
		return "", err
	}
	return result, nil
}

// RenderTemplateFile is a shortcut and renders a template file directly.
func (set *TemplateSet) RenderTemplateFile(fn string, ctx Context) (string, error) {
	tpl := Must(set.FromFile(fn))
	result, err := tpl.Execute(ctx)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (set *TemplateSet) logf(format string, args ...any) {
	if set.Debug {
		logger.Printf(fmt.Sprintf("[template set: %s] %s", set.name, format), args...)
	}
}

// Logging function (internally used)
func logf(format string, items ...any) {
	if debug {
		logger.Printf(format, items...)
	}
}

var (
	debug  bool // internal debugging
	logger = log.New(os.Stdout, "[pongo2] ", log.LstdFlags|log.Lshortfile)

	// DefaultLoader allows the default un-sandboxed access to the local file
	// system and is being used by the DefaultSet.
	DefaultLoader = MustNewLocalFileSystemLoader("")

	// DefaultSet is a set created for you for convinience reasons.
	DefaultSet = newSet("default", DefaultLoader)

	// Methods on the default set
	FromString           = DefaultSet.FromString
	FromBytes            = DefaultSet.FromBytes
	FromFile             = DefaultSet.FromFile
	FromCache            = DefaultSet.FromCache
	RenderTemplateString = DefaultSet.RenderTemplateString
	RenderTemplateFile   = DefaultSet.RenderTemplateFile
	ReplaceTag           = DefaultSet.ReplaceTag
	RegisterTag          = DefaultSet.RegisterTag
	ReplaceFilter        = DefaultSet.ReplaceFilter
	RegisterFilter       = DefaultSet.RegisterFilter
	ApplyFilter          = DefaultSet.ApplyFilter
	MustApplyFilter      = DefaultSet.MustApplyFilter

	// Globals for the default set
	Globals = DefaultSet.Globals
)
