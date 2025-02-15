package pongo2

// props tag
// Usage:
// {%- props key1=value1, key2=value2, key3=value3 -%}
// {%- props key1=value1, key2, key3 -%}

type attributes struct {
	keyValues map[string]IEvaluator
}

type tagsPropsNode struct {
	keyValues map[string]IEvaluator
}

func (node *tagsPropsNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	for key, value := range node.keyValues {
		// If the key is not set in the public context,
		// it means the component was called without the prop.
		// In this case, we set the default value.
		if ctx.Public[key] == nil {
			val, err := value.Evaluate(ctx)
			if err != nil {
				return err
			}
			// When setting a private context, the specified value will override the existing one.
			ctx.Private[key] = val
		}
	}

	return nil
}

func tagPropsParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	propsNode := &tagsPropsNode{
		keyValues: make(map[string]IEvaluator),
	}

	propKeys := make([]string, 0)

	// Parse arguments
	for arguments.Remaining() > 0 {
		// Retrieve the key
		keyToken := arguments.MatchType(TokenIdentifier)
		if keyToken == nil {
			return nil, arguments.Error("Expected a key (identifier).", nil)
		}
		key := keyToken.Val

		// Check for `=` and retrieve the value if present
		var value IEvaluator
		if arguments.Match(TokenSymbol, "=") != nil {
			var err *Error
			value, err = arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
		} else {
			// If `=` is not present, value is nil
			value = nil
		}

		// Add the key and value to the map
		if value != nil {
			propsNode.keyValues[key] = value
		}
		// Add the key to the list of keys
		propKeys = append(propKeys, key)

		// If the next token is a comma, consume it
		arguments.Match(TokenSymbol, ",") // No need to break, just consume
	}

	// Check for syntax errors
	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Unexpected token after props arguments.", nil)
	}

	// save defined props
	doc.template.props = propKeys

	return propsNode, nil
}

func init() {
	RegisterTag("props", tagPropsParser)
}
