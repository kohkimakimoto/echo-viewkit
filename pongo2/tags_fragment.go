package pongo2

type tagFragmentNode struct {
	name    string
	wrapper *NodeWrapper
}

func (node *tagFragmentNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	return node.wrapper.Execute(ctx, writer)
}

func tagFragmentParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	fragmentNode := &tagFragmentNode{}

	nameToken := arguments.MatchType(TokenString)
	if nameToken == nil {
		return nil, arguments.Error("fragment tag needs at least a string as name.", nil)
	}
	fragmentNode.name = nameToken.Val

	wrapper, endtagargs, err := doc.WrapUntilTag("endfragment")
	if err != nil {
		return nil, err
	}
	if endtagargs.Count() > 0 {
		return nil, endtagargs.Error("Arguments not allowed here.", nil)
	}
	fragmentNode.wrapper = wrapper

	doc.template.fragments[fragmentNode.name] = wrapper
	return fragmentNode, nil
}

func init() {
	RegisterTag("fragment", tagFragmentParser)
}
