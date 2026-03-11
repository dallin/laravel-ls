package provider

import ts "github.com/tree-sitter/go-tree-sitter"

type InlayHint struct {
	Position ts.Point
	Label    string
}

type InlayHintPublish func(InlayHint)

type InlayHintContext struct {
	BaseContext
	Publish InlayHintPublish
}

type InlayHintProvider interface {
	ResolveInlayHints(InlayHintContext)
}
