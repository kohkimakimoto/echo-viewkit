package pongo2

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

type ComponentExecutionContext struct {
	EchoContext echo.Context
	Data        Context
}

func (c *ComponentExecutionContext) Bind(out any) error {
	return UnmarshalContext(c.Data, out)
}

func (c *ComponentExecutionContext) Update(data any) error {
	ctx, err := MarshalContext(data)
	if err != nil {
		return err
	}
	c.Data = c.Data.Update(ctx)
	return nil
}

func (c *ComponentExecutionContext) Set(key string, value any) {
	c.Data[key] = value
}

func (c *ComponentExecutionContext) Get(key string) any {
	return c.Data[key]
}

func (c *ComponentExecutionContext) Delete(key string) {
	delete(c.Data, key)
}

// Default sets the default value for the key
func (c *ComponentExecutionContext) Default(key string, value any) {
	if _, ok := c.Data[key]; !ok {
		c.Data[key] = value
	}
}

// Defaults sets the default values
func (c *ComponentExecutionContext) Defaults(data any) error {
	ctx, err := MarshalContext(data)
	if err != nil {
		return err
	}
	for key, value := range ctx {
		c.Default(key, value)
	}
	return nil
}

func (c *ComponentExecutionContext) Attributes() *Attributes {
	return c.Data["attributes"].(*Attributes)
}

// component is an internal representation of a component.
type component struct {
	Name           string
	TemplateFile   string
	TemplateString string
	Props          []string
	Setup          func(*ComponentExecutionContext) error
}

type componentSet struct {
	// TemplateSetFS is a root directory for template files.
	TemplateSetFS                fs.FS
	DefaultTemplateFileExtension string
	// registered components
	components map[string]*component
}

func newComponentSet() *componentSet {
	return &componentSet{
		TemplateSetFS: nil,
		components:    make(map[string]*component),
	}
}

type Component struct {
	Name         string
	TemplateFile string
	Props        []string
	Setup        func(*ComponentExecutionContext) error
}

func (set *componentSet) RegisterComponent(comp *Component) {
	set.components[comp.Name] = &component{
		Name:           comp.Name,
		TemplateFile:   comp.TemplateFile,
		TemplateString: "",
		Props:          comp.Props,
		Setup:          comp.Setup,
	}
}

type AnonymousComponent struct {
	Name         string
	TemplateFile string
}

func (set *componentSet) RegisterAnonymousComponent(comp *AnonymousComponent) {
	set.components[comp.Name] = &component{
		Name:           comp.Name,
		TemplateFile:   comp.TemplateFile,
		TemplateString: "",
		Props:          nil,
		Setup:          nil,
	}
}

type AnonymousComponentsDirectory struct {
	Prefix string
	Dir    string
}

// RegisterTemplateFileComponentsDirectory registers all template files in the specified directory as components.
func (set *componentSet) RegisterTemplateFileComponentsDirectory(compDir *AnonymousComponentsDirectory) error {
	return fs.WalkDir(set.TemplateSetFS, compDir.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			templateFile := strings.TrimPrefix(path, "/")
			name := strings.TrimPrefix(strings.TrimPrefix(templateFile, compDir.Dir), "/")

			// Remove the default template extension from the name, if it is set
			if set.DefaultTemplateFileExtension != "" && filepath.Ext(name) == set.DefaultTemplateFileExtension {
				name = name[:len(name)-len(set.DefaultTemplateFileExtension)]
			}

			// Replace the path separator with a dot
			name = strings.ReplaceAll(filepath.ToSlash(name), "/", ".")
			if compDir.Prefix != "" {
				name = compDir.Prefix + name
			}

			set.RegisterAnonymousComponent(&AnonymousComponent{
				Name:         name,
				TemplateFile: templateFile,
			})
		}

		return nil
	})
}

type InlineComponent struct {
	Name           string
	TemplateString string
	Props          []string
	Setup          func(*ComponentExecutionContext) error
}

func (set *componentSet) RegisterInlineComponent(comp *InlineComponent) {
	set.components[comp.Name] = &component{
		Name:           comp.Name,
		TemplateFile:   "",
		TemplateString: comp.TemplateString,
		Props:          comp.Props,
		Setup:          comp.Setup,
	}
}

type HeadlessComponent struct {
	Name  string
	Props []string
	Setup func(*ComponentExecutionContext) error
}

func (set *componentSet) RegisterHeadlessComponent(comp *HeadlessComponent) {
	set.components[comp.Name] = &component{
		Name:           comp.Name,
		TemplateFile:   "",
		TemplateString: "{{ slot }}",
		Props:          comp.Props,
		Setup:          comp.Setup,
	}
}

func (set *componentSet) resolveComponent(name string) *component {
	if component, ok := set.components[name]; ok {
		return component
	}
	return nil
}

type tagComponentNode struct {
	id        string
	tpl       *Template
	component *component
	attrs     []*tagComponentAttribute
	data      map[string]IEvaluator
	slots     []*componentSlot
	slotData  *slotData
}

type tagComponentAttribute struct {
	name string
	expr IEvaluator
}

type componentSlot struct {
	Name    string
	wrapper *NodeWrapper
}

type slotData struct {
	name string
	keys []*slotDataKey
}

type slotDataKey struct {
	name  string
	alias string
}

var (
	// Regular expression to validate the entire input format
	validContentRegex = regexp.MustCompile(`^(\w+(:\s*\w+)?)(,\s*\w+(:\s*\w+)?)*$`)
	// Regular expression to match each key and alias pair
	slotDataRegex = regexp.MustCompile(`(\w+)(?::\s*(\w+))?`)
)

func parseSlotDataExpr(expr string) (*slotData, error) {
	// Trim whitespace from the input
	expr = strings.TrimSpace(expr)

	// Ensure the input is enclosed in braces
	if !strings.HasPrefix(expr, "{") || !strings.HasSuffix(expr, "}") {
		// Return a single name slot data if the input is not enclosed in braces
		return &slotData{name: expr, keys: make([]*slotDataKey, 0)}, nil
	}

	// Remove braces from the input
	content := strings.Trim(expr, "{}")
	content = strings.TrimSpace(content)

	// Return an empty slot data if the content is empty
	if content == "" {
		return &slotData{name: "", keys: make([]*slotDataKey, 0)}, nil
	}

	// Validate the content format
	if !validContentRegex.MatchString(content) {
		return nil, fmt.Errorf("invalid format: contains invalid characters or structure")
	}

	matches := slotDataRegex.FindAllStringSubmatch(content, -1)

	var keys []*slotDataKey
	for _, match := range matches {
		// Extract the key name and alias
		name := match[1]
		alias := match[2]

		// Append the parsed key to the keys
		keys = append(keys, &slotDataKey{name: name, alias: alias})
	}

	return &slotData{
		name: "",
		keys: keys,
	}, nil
}

var ErrNoComponentContent = errors.New("no component content")

func (node *tagComponentNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	// create component scope new context
	newCtx := make(Context)

	// copy all data into the context
	for key, value := range node.data {
		val, err := value.Evaluate(ctx)
		if err != nil {
			return err
		}
		newCtx[key] = val
	}

	// create attributes
	var attrPairs [][2]string
	for _, attr := range node.attrs {
		val, err := attr.expr.Evaluate(ctx)
		if err != nil {
			return err
		}
		attrPairs = append(attrPairs, [2]string{attr.name, val.String()})
	}
	newCtx["attributes"] = newAttributes(attrPairs)

	// execute the component Setup function
	if node.component.Setup != nil {
		err := node.component.Setup(&ComponentExecutionContext{
			EchoContext: ctx.echoContext,
			Data:        newCtx,
		})
		if err != nil {
			// if the component action returns ErrNoComponentContent, do nothing.
			if errors.Is(err, ErrNoComponentContent) {
				return nil
			}
			return ctx.OrigError(err, nil)
		}
	}

	// copy shared context keys
	for _, key := range ctx.template.set.SharedContextKeys {
		if value, ok := ctx.Public[key]; ok {
			newCtx[key] = value
		}
	}

	// execute the slots
	for _, slot := range node.slots {
		slotCtx := NewChildExecutionContext(ctx)
		if node.slotData != nil {
			// expose the component data to the slot
			if node.slotData.name != "" {
				slotCtx.Private[node.slotData.name] = newCtx
			} else {
				// extract specific parameters directly from the component data
				for _, key := range node.slotData.keys {
					if key.alias == "" {
						slotCtx.Private[key.name] = newCtx[key.name]
					} else {
						slotCtx.Private[key.alias] = newCtx[key.name]
					}
				}
			}
		}

		var b bytes.Buffer
		if err := slot.wrapper.Execute(slotCtx, &b); err != nil {
			return err
		}
		newCtx[slot.Name] = AsSafeValue(strings.TrimSpace(b.String()))
	}

	// Execute the component template
	err := node.tpl.ExecuteWriterWithEchoContext(newCtx, writer, ctx.echoContext)
	if err != nil {
		return err.(*Error)
	}

	return nil
}

// The component tag is like the following:
// {% component "alert" withAttrs "message"="text" "type"=type %}

func tagComponentParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	componentNode := &tagComponentNode{
		attrs:    make([]*tagComponentAttribute, 0),
		data:     make(map[string]IEvaluator),
		slots:    make([]*componentSlot, 0),
		slotData: nil,
	}

	componentNameToken := arguments.MatchType(TokenString)
	if componentNameToken == nil {
		return nil, arguments.Error("component tag needs a component name as first argument.", nil)
	}
	componentName := componentNameToken.Val
	comp := doc.template.set.ComponentSet.resolveComponent(componentName)
	if comp == nil {
		return nil, arguments.Error(fmt.Sprintf("component '%s' can not be resolved.", componentName), nil)
	}
	componentNode.component = comp

	if comp.TemplateFile != "" {
		// Load the template from the file system
		tpl, err := doc.template.set.FromFile(comp.TemplateFile)
		if err != nil {
			return nil, err.(*Error)
		}
		componentNode.tpl = tpl
	} else if comp.TemplateString != "" {
		// Load the template from the string
		tpl, err := doc.template.set.FromString(comp.TemplateString)
		if err != nil {
			return nil, err.(*Error)
		}
		componentNode.tpl = tpl
	} else {
		return nil, arguments.Error(fmt.Sprintf("component '%s' has no template.", componentName), nil)
	}

	// get props definition:
	var props []string
	if len(comp.Props) > 0 {
		// get from component
		props = comp.Props
	} else if len(componentNode.tpl.props) > 0 {
		// get from template {% props %} tag
		props = componentNode.tpl.props
	}
	propsMap := make(map[string]bool)
	for _, prop := range props {
		propsMap[prop] = true
	}

	// After having parsed the component name we're going to parse the additional options

	// data (slot-data) property
	if arguments.Match(TokenIdentifier, "slotData") != nil {
		if arguments.Match(TokenSymbol, "=") == nil {
			return nil, arguments.Error("Expected '='.", nil)
		}
		slotDataExpr := arguments.MatchType(TokenString)
		if slotDataExpr == nil {
			return nil, arguments.Error("slotData (slot-data in with an HTML-like syntax) property needs value.", nil)
		}

		sd, err := parseSlotDataExpr(slotDataExpr.Val)
		if err != nil {
			return nil, arguments.Error(fmt.Sprintf("slotData (slot-data in with an HTML-like syntax) value is invalid: %v", err), nil)
		}
		componentNode.slotData = sd
	}

	// with options
	if arguments.Match(TokenIdentifier, "withAttrs") != nil {
		for arguments.Remaining() > 0 {
			// We have at least one "key"=expr pair (because of starting "withAttrs")
			keyToken := arguments.MatchType(TokenString)
			if keyToken == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			if arguments.Match(TokenSymbol, "=") == nil {
				return nil, arguments.Error("Expected '='.", nil)
			}
			valueExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}

			// Check if the key is a prop or a fallthrough attribute
			if propsMap[keyToken.Val] {
				// the key is a prop
				componentNode.data[keyToken.Val] = valueExpr
			} else {
				// the key is a fallthrough attribute
				componentNode.attrs = append(
					componentNode.attrs,
					&tagComponentAttribute{
						name: keyToken.Val,
						expr: valueExpr,
					},
				)
			}
		}
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed 'component'-tag arguments.", nil)
	}

	for {
		wrapper, tagArgs, err := doc.WrapUntilTag("slot", "endcomponent")
		if err != nil {
			return nil, err
		}

		if wrapper.Endtag == "slot" {
			slotNameToken := tagArgs.MatchType(TokenString)
			if slotNameToken == nil {
				return nil, tagArgs.Error("slot tag needs a slot name as first argument.", nil)
			}

			wrapper, tagArgs, err := doc.WrapUntilTag("endslot")
			if err != nil {
				return nil, err
			}
			if tagArgs.Count() > 0 {
				return nil, tagArgs.Error("Arguments not allowed here.", nil)
			}
			componentNode.slots = append(componentNode.slots, &componentSlot{Name: slotNameToken.Val, wrapper: wrapper})
		} else if wrapper.Endtag == "endcomponent" {
			if tagArgs.Count() > 0 {
				return nil, tagArgs.Error("Arguments not allowed here.", nil)
			}
			componentNode.slots = append(componentNode.slots, &componentSlot{Name: "slot", wrapper: wrapper})
			break
		}
	}

	return componentNode, nil
}

func init() {
	RegisterTag("component", tagComponentParser)
}
