package provider

import ts "github.com/tree-sitter/go-tree-sitter"

type InlayHint struct {
	Position ts.Point
	Label    string
}

type InlayHintPublish func(InlayHint)

type InlayHintContext struct {
	BaseContext
	// Range is the visible document range for which hints were requested.
	// Providers may use this to skip work outside the visible area.
	Range   ts.Range
	Publish InlayHintPublish
}

// Interface that providers that support inlay hints can implement.
type InlayHintProvider interface {
	ResolveInlayHints(InlayHintContext)
}
