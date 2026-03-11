package route

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/laravel-ls/laravel-ls/file"
	"github.com/laravel-ls/laravel-ls/laravel/providers/route/queries"
	"github.com/laravel-ls/laravel-ls/project"
	"github.com/laravel-ls/laravel-ls/provider"
	"github.com/laravel-ls/laravel-ls/treesitter/php"
	"github.com/laravel-ls/laravel-ls/utils/repository"
	"github.com/laravel-ls/protocol"
	log "github.com/sirupsen/logrus"
)

type Provider struct {
	rootPath string
	project  *project.Project

	mu         sync.Mutex
	routeCache repository.RouteRepository
	routeGen   uint64 // incremented on invalidation; prevents stale PHP results from overwriting a cleared cache
}

func NewProvider() *Provider {
	return &Provider{}
}

// OnFileSaved invalidates the route cache when a file under routes/ is saved.
// Returns a channel that closes when the cache has been re-warmed via a fresh
// PHP call, or nil if the file was not in the routes directory.
func (p *Provider) OnFileSaved(filename string) <-chan struct{} {
	routesDir := path.Join(p.rootPath, "routes") + "/"
	if !strings.HasPrefix(filename, routesDir) {
		return nil
	}

	if p.project == nil {
		return nil
	}

	p.mu.Lock()
	p.routeCache = nil
	p.routeGen++
	gen := p.routeGen
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		repo, err := p.project.Routes()
		if err != nil {
			log.WithError(err).Debug("routes: OnFileSaved pre-warm failed")
			return
		}
		p.mu.Lock()
		if p.routeGen == gen {
			log.WithField("routes", len(repo)).WithField("gen", gen).Debug("routes: pre-warm complete, storing in cache")
			p.routeCache = repo
		}
		p.mu.Unlock()
	}()
	return done
}

// routes returns the route repository, cached in memory and invalidated on
// routes file save. The first call after startup or invalidation spawns a
// PHP process to load the routes; subsequent calls return the cached result.
func (p *Provider) routes(ctx provider.BaseContext) (repository.RouteRepository, error) {
	p.mu.Lock()
	cache, gen := p.routeCache, p.routeGen
	p.mu.Unlock()

	if cache != nil {
		return cache, nil
	}

	repo, err := ctx.Project.Routes()
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	// Only write back if no invalidation happened while PHP was running.
	if p.routeGen == gen {
		log.WithField("routes", len(repo)).WithField("gen", gen).Debug("routes: PHP call complete, storing in cache")
		p.routeCache = repo
	} else {
		log.WithField("gen_expected", gen).WithField("gen_current", p.routeGen).Debug("routes: PHP result discarded (invalidated during call)")
	}
	p.mu.Unlock()

	return repo, nil
}

func (p *Provider) Register(manager *provider.Manager) {
	manager.Register(file.TypePHP, p)
}

func (p *Provider) Init(ctx provider.InitContext) {
	p.rootPath = ctx.RootPath
	p.project = ctx.Project
}

func (p *Provider) Hover(ctx provider.HoverContext) {
	node := queries.RouteCalls(ctx.File).At(ctx.Position)
	if node == nil {
		return
	}

	text := php.GetStringContent(node, ctx.File.Src)
	if len(text) < 1 {
		return
	}

	repo, err := p.routes(ctx.BaseContext)
	if err != nil {
		ctx.Logger.WithError(err).Warn("failed to get repo")
		return
	}

	route, ok := repo.Get(text)
	if !ok {
		return
	}

	ctx.Publish(provider.Hover{
		Content: formatHoverContent(p.rootPath, route),
	})
}

func (p *Provider) ResolveCompletion(ctx provider.CompletionContext) {
	node := queries.RouteCalls(ctx.File).At(ctx.Position)
	if node == nil {
		return
	}

	text := php.GetStringContent(node, ctx.File.Src)

	repo, err := p.routes(ctx.BaseContext)
	if err != nil {
		ctx.Logger.WithError(err).Warn("failed to get repo")
		return
	}

	for key, meta := range repo.Find(text) {
		ctx.Publish(formatCompetionItem(key, meta))
	}
}

func (p *Provider) ResolveDefinition(ctx provider.DefinitionContext) {
	node := queries.RouteCalls(ctx.File).At(ctx.Position)
	if node == nil {
		return
	}

	route := php.GetStringContent(node, ctx.File.Src)
	if len(route) < 1 {
		return
	}

	repo, err := p.routes(ctx.BaseContext)
	if err != nil {
		ctx.Logger.WithError(err).Warn("failed to get repo")
		return
	}

	if meta, found := repo.Get(route); found {
		ctx.Publish(formatLocation(meta))
	}
}

func (p *Provider) Diagnostic(ctx provider.DiagnosticContext) {
	node := queries.RouteCalls(ctx.File)
	if len(node) < 1 {
		return
	}

	repo, err := p.routes(ctx.BaseContext)
	if err != nil {
		ctx.Logger.WithError(err).Warn("failed to get repo")
		return
	}

	for _, capture := range node {
		text := php.GetStringContent(&capture.Node, ctx.File.Src)
		if len(text) < 1 || repo.Exists(text) {
			continue
		}

		// Follow format from:
		// https://github.com/laravel/vs-code-extension/blob/v1.0.11/src/features/route.ts#L137-L142
		// https://github.com/laravel/vs-code-extension/blob/main/src/diagnostic/index.ts#L3-L14
		ctx.Publish(provider.Diagnostic{
			Range:    capture.Node.Range(),
			Severity: protocol.DiagnosticSeverityWarning,
			Message:  fmt.Sprintf("Route [%s] not found", text),
		})
	}
}

func (p *Provider) ResolveCodeAction(ctx provider.CodeActionContext) {
	nodes := queries.RouteCalls(ctx.File).In(ctx.Range)
	if len(nodes) < 1 {
		return
	}

	repo, err := p.routes(ctx.BaseContext)
	if err != nil {
		ctx.Logger.WithError(err).Warn("failed to get repo")
		return
	}

	routeFilename := path.Join(p.rootPath, "routes/web.php")
	routeFile, err := ctx.FileCache.Open(routeFilename)
	if err != nil {
		ctx.Logger.WithError(err).Warn("failed to parse routes/web.php file")
		return
	}

	for _, node := range nodes {
		text := php.GetStringContent(node, ctx.File.Src)
		if len(text) < 1 {
			return
		}

		if _, found := repo.Get(text); !found {
			uri := protocol.DocumentURI("file://" + routeFilename)
			line := routeFile.Tree.Root().EndPosition().Row

			code := fmt.Sprintf(routeTemplate, text)

			ctx.Publish(codeAction(uri, "Add to routes file (web.php)", line, code))
		}
	}
}
