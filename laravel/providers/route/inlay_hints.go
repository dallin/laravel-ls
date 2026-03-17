package route

import (
	"fmt"
	"strings"

	"github.com/laravel-ls/laravel-ls/parser"
	"github.com/laravel-ls/laravel-ls/provider"
	"github.com/laravel-ls/laravel-ls/treesitter"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type controllerInfo struct {
	FQN     string              // e.g. "App\Http\Controllers\HomeController"
	Methods map[string]ts.Point // method name -> start position
}

func parseControllerInfo(file *parser.File) *controllerInfo {
	root := file.Tree.Root()

	var namespace string
	var className string
	methods := map[string]ts.Point{}

	for i := uint(0); i < root.NamedChildCount(); i++ {
		child := root.NamedChild(i)
		switch child.Kind() {
		case "namespace_definition":
			if nameNode := treesitter.NamedChildOfKind(child, "namespace_name"); nameNode != nil {
				namespace = nameNode.Utf8Text(file.Src)
			}
		case "class_declaration":
			if nameNode := treesitter.NamedChildOfKind(child, "name"); nameNode != nil {
				className = nameNode.Utf8Text(file.Src)
			}
			body := treesitter.NamedChildOfKind(child, "declaration_list")
			if body == nil {
				continue
			}
			for j := uint(0); j < body.NamedChildCount(); j++ {
				member := body.NamedChild(j)
				if member.Kind() != "method_declaration" {
					continue
				}
				if methodName := treesitter.NamedChildOfKind(member, "name"); methodName != nil {
					methods[methodName.Utf8Text(file.Src)] = member.StartPosition()
				}
			}
		}
	}

	if className == "" {
		return nil
	}

	fqn := className
	if namespace != "" {
		fqn = namespace + "\\" + className
	}

	return &controllerInfo{
		FQN:     fqn,
		Methods: methods,
	}
}

func (p *Provider) ResolveInlayHints(ctx provider.InlayHintContext) {
	repo, err := p.routes()
	if err != nil {
		return
	}

	info := parseControllerInfo(ctx.File)
	if info == nil {
		return
	}

	fqnParts := strings.Split(info.FQN, "\\")
	classBaseName := fqnParts[len(fqnParts)-1]

	for _, route := range repo {
		parts := strings.SplitN(route.Action, "@", 2)
		// Match by full FQN (e.g. "App\Http\Controllers\HomeController") or by unqualified class name (e.g. "HomeController").
		if parts[0] != info.FQN && parts[0] != classBaseName {
			continue
		}

		methodName := "__invoke"
		if len(parts) == 2 {
			methodName = parts[1]
		}

		pos, ok := info.Methods[methodName]
		if !ok {
			continue
		}

		ctx.Publish(provider.InlayHint{
			Position: pos,
			Label:    fmt.Sprintf("[%s /%s]", route.Method, strings.TrimPrefix(route.URI, "/")),
		})
	}
}
