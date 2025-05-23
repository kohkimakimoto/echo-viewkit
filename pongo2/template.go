package pongo2

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"strings"
)

type TemplateWriter interface {
	io.Writer
	WriteString(string) (int, error)
}

type templateWriter struct {
	w io.Writer
}

func (tw *templateWriter) WriteString(s string) (int, error) {
	return tw.w.Write([]byte(s))
}

func (tw *templateWriter) Write(b []byte) (int, error) {
	return tw.w.Write(b)
}

type Template struct {
	id string

	set *TemplateSet

	// Input
	isTplString bool
	name        string
	tpl         string
	size        int

	// Calculation
	tokens []*Token
	parser *Parser

	// first come, first serve (it's important to not override existing entries in here)
	level          int
	parent         *Template
	child          *Template
	blocks         map[string]*NodeWrapper
	exportedMacros map[string]*tagMacroNode

	// Defined props of a component
	props []string

	// fragments
	fragments map[string]*NodeWrapper

	// Output
	root *nodeDocument

	// Options allow you to change the behavior of template-engine.
	// You can change the options before calling the Execute method.
	Options *Options
}

func newTemplateString(set *TemplateSet, tpl []byte) (*Template, error) {
	return newTemplate(set, "<string>", true, tpl)
}

func newTemplate(set *TemplateSet, name string, isTplString bool, tpl []byte) (*Template, error) {
	strTpl := string(tpl)

	// Create the template
	t := &Template{
		id:             genRandomId(),
		set:            set,
		isTplString:    isTplString,
		name:           name,
		tpl:            strTpl,
		size:           len(strTpl),
		blocks:         make(map[string]*NodeWrapper),
		exportedMacros: make(map[string]*tagMacroNode),
		props:          make([]string, 0),
		fragments:      make(map[string]*NodeWrapper),
		Options:        newOptions(),
	}
	// Copy all settings from another Options.
	t.Options.Update(set.Options)

	// Tokenize it
	tokens, err := lex(name, strTpl)
	if err != nil {
		return nil, err
	}
	t.tokens = tokens

	// For debugging purposes, show all tokens:
	/*for i, t := range tokens {
		fmt.Printf("%3d. %s\n", i, t)
	}*/

	// Parse it
	err = t.parse()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (tpl *Template) newContextForExecution(context Context) (*Template, *ExecutionContext, error) {
	return tpl.newContextForExecutionWithEchoContext(context, nil)
}

func (tpl *Template) newContextForExecutionWithEchoContext(context Context, eCtx echo.Context) (*Template, *ExecutionContext, error) {
	if tpl.Options.TrimBlocks || tpl.Options.LStripBlocks {
		// Issue #94 https://github.com/flosch/pongo2/issues/94
		// If an application configures pongo2 template to trim_blocks,
		// the first newline after a template tag is removed automatically (like in PHP).
		prev := &Token{
			Typ: TokenHTML,
			Val: "\n",
		}

		for _, t := range tpl.tokens {
			if tpl.Options.LStripBlocks {
				if prev.Typ == TokenHTML && t.Typ != TokenHTML && t.Val == "{%" {
					prev.Val = strings.TrimRight(prev.Val, "\t ")
				}
			}

			if tpl.Options.TrimBlocks {
				if prev.Typ != TokenHTML && t.Typ == TokenHTML && prev.Val == "%}" {
					if len(t.Val) > 0 && t.Val[0] == '\n' {
						t.Val = t.Val[1:len(t.Val)]
					}
				}
			}

			prev = t
		}
	}

	// Determine the parent to be executed (for template inheritance)
	parent := tpl
	for parent.parent != nil {
		parent = parent.parent
	}

	// Create context if none is given
	newContext := make(Context)
	newContext.Update(tpl.set.Globals)

	if context != nil {
		newContext.Update(context)

		if len(newContext) > 0 {
			// Check for context name syntax
			err := newContext.checkForValidIdentifiers()
			if err != nil {
				return parent, nil, err
			}

			// Check for clashes with macro names
			for k := range newContext {
				_, has := tpl.exportedMacros[k]
				if has {
					return parent, nil, &Error{
						Filename:  tpl.name,
						Sender:    "execution",
						OrigError: fmt.Errorf("context key name '%s' clashes with macro '%s'", k, k),
					}
				}
			}
		}
	}

	// Create operational context
	ctx := newExecutionContextWithEchoContext(parent, newContext, eCtx)

	return parent, ctx, nil
}

func (tpl *Template) execute(context Context, writer TemplateWriter) error {
	return tpl.executeWithEchoContext(context, writer, nil)
}

func (tpl *Template) executeWithEchoContext(context Context, writer TemplateWriter, eCtx echo.Context) error {
	parent, ctx, err := tpl.newContextForExecutionWithEchoContext(context, eCtx)
	if err != nil {
		return err
	}

	// Run the selected document
	if err := parent.root.Execute(ctx, writer); err != nil {
		return err
	}

	return nil
}

func (tpl *Template) executeFragmentWithEchoContext(context Context, fragmentName string, writer TemplateWriter, eCtx echo.Context) error {
	parent, ctx, err := tpl.newContextForExecutionWithEchoContext(context, eCtx)
	if err != nil {
		return err
	}

	fragment, ok := parent.fragments[fragmentName]
	if !ok {
		return fmt.Errorf("fragment '%s' not found", fragmentName)
	}

	// Run the selected fragment
	if err := fragment.Execute(ctx, writer); err != nil {
		return err
	}

	return nil
}

func (tpl *Template) newTemplateWriterAndExecute(context Context, writer io.Writer) error {
	return tpl.execute(context, &templateWriter{w: writer})
}

func (tpl *Template) newBufferAndExecute(context Context) (*bytes.Buffer, error) {
	return tpl.newBufferAndExecuteWithEchoContext(context, nil)
}

func (tpl *Template) newBufferAndExecuteWithEchoContext(context Context, eCtx echo.Context) (*bytes.Buffer, error) {
	// Create output buffer
	// We assume that the rendered template will be 30% larger
	buffer := bytes.NewBuffer(make([]byte, 0, int(float64(tpl.size)*1.3)))
	if err := tpl.executeWithEchoContext(context, buffer, eCtx); err != nil {
		return nil, err
	}
	return buffer, nil
}

func (tpl *Template) newBufferAndExecuteFragmentWithEchoContext(context Context, fragmentName string, eCtx echo.Context) (*bytes.Buffer, error) {
	// Create output buffer
	// We assume that the rendered template will be 30% larger
	buffer := bytes.NewBuffer(make([]byte, 0, int(float64(tpl.size)*1.3)))
	if err := tpl.executeFragmentWithEchoContext(context, fragmentName, buffer, eCtx); err != nil {
		return nil, err
	}
	return buffer, nil
}

// Executes the template with the given context and writes to writer (io.Writer)
// on success. Context can be nil. Nothing is written on error; instead the error
// is being returned.
func (tpl *Template) ExecuteWriter(context Context, writer io.Writer) error {
	return tpl.ExecuteWriterWithEchoContext(context, writer, nil)
}

func (tpl *Template) ExecuteWriterWithEchoContext(context Context, writer io.Writer, eCtx echo.Context) error {
	buf, err := tpl.newBufferAndExecuteWithEchoContext(context, eCtx)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(writer)
	if err != nil {
		return err
	}
	return nil
}

func (tpl *Template) ExecuteFragmentWriterWithEchoContext(context Context, fragmentName string, writer io.Writer, eCtx echo.Context) error {
	buf, err := tpl.newBufferAndExecuteFragmentWithEchoContext(context, fragmentName, eCtx)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(writer)
	if err != nil {
		return err
	}
	return nil
}

// Same as ExecuteWriter. The only difference between both functions is that
// this function might already have written parts of the generated template in the
// case of an execution error because there's no intermediate buffer involved for
// performance reasons. This is handy if you need high performance template
// generation or if you want to manage your own pool of buffers.
func (tpl *Template) ExecuteWriterUnbuffered(context Context, writer io.Writer) error {
	return tpl.newTemplateWriterAndExecute(context, writer)
}

// Executes the template and returns the rendered template as a []byte
func (tpl *Template) ExecuteBytes(context Context) ([]byte, error) {
	// Execute template
	buffer, err := tpl.newBufferAndExecute(context)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Executes the template and returns the rendered template as a string
func (tpl *Template) Execute(context Context) (string, error) {
	// Execute template
	buffer, err := tpl.newBufferAndExecute(context)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (tpl *Template) ExecuteBlocks(context Context, blocks []string) (map[string]string, error) {
	var parents []*Template
	result := make(map[string]string)

	parent := tpl
	for parent != nil {
		// We only want to execute the template if it has a block we want
		for _, block := range blocks {
			if _, ok := tpl.blocks[block]; ok {
				parents = append(parents, parent)
				break
			}
		}
		parent = parent.parent
	}

	for _, t := range parents {
		var buffer *bytes.Buffer
		var ctx *ExecutionContext
		var err error
		for _, blockName := range blocks {
			if _, ok := result[blockName]; ok {
				continue
			}
			if blockWrapper, ok := t.blocks[blockName]; ok {
				// assign the buffer if we haven't done so
				if buffer == nil {
					buffer = bytes.NewBuffer(make([]byte, 0, int(float64(t.size)*1.3)))
				}
				// assign the context if we haven't done so
				if ctx == nil {
					_, ctx, err = t.newContextForExecution(context)
					if err != nil {
						return nil, err
					}
				}
				bErr := blockWrapper.Execute(ctx, buffer)
				if bErr != nil {
					return nil, bErr
				}
				result[blockName] = buffer.String()
				buffer.Reset()
			}
		}
		// We have found all blocks
		if len(blocks) == len(result) {
			break
		}
	}

	return result, nil
}
