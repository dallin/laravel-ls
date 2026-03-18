package route

import (
	"testing"

	"github.com/laravel-ls/laravel-ls/file"
	"github.com/laravel-ls/laravel-ls/parser"
	"github.com/laravel-ls/laravel-ls/provider"
	"github.com/laravel-ls/laravel-ls/utils/repository"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func mustParsePHP(t *testing.T, src string) *parser.File {
	t.Helper()
	f, err := parser.Parse([]byte(src), file.TypePHP)
	require.NoError(t, err)
	return f
}

func TestParseControllerInfo(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		expectNil bool
		fqn       string
		methods   []string
	}{
		{
			name: "controller with namespace",
			src: `<?php
namespace App\Http\Controllers;

class HomeController
{
    public function index() {}
    public function show() {}
}`,
			fqn:     `App\Http\Controllers\HomeController`,
			methods: []string{"index", "show"},
		},
		{
			name: "controller without namespace",
			src: `<?php
class UserController
{
    public function index() {}
}`,
			fqn:     "UserController",
			methods: []string{"index"},
		},
		{
			name: "invokable controller",
			src: `<?php
namespace App\Http\Controllers;

class ShowDashboard
{
    public function __invoke() {}
}`,
			fqn:     `App\Http\Controllers\ShowDashboard`,
			methods: []string{"__invoke"},
		},
		{
			name:      "non-class PHP file",
			src:       `<?php echo "hello";`,
			expectNil: true,
		},
		{
			name:      "route file with closures",
			src:       `<?php use Illuminate\Support\Facades\Route; Route::get('/', function () { return view('welcome'); });`,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := mustParsePHP(t, tt.src)
			info := parseControllerInfo(f)

			if tt.expectNil {
				require.Nil(t, info)
				return
			}

			require.NotNil(t, info)
			require.Equal(t, tt.fqn, info.FQN)
			for _, method := range tt.methods {
				_, ok := info.Methods[method]
				require.True(t, ok, "expected method %q in controller %q", method, tt.fqn)
			}
			require.Len(t, info.Methods, len(tt.methods))
		})
	}
}

func resolveHints(t *testing.T, p *Provider, src string) []provider.InlayHint {
	t.Helper()
	f := mustParsePHP(t, src)
	var hints []provider.InlayHint
	p.ResolveInlayHints(provider.InlayHintContext{
		BaseContext: provider.BaseContext{
			Logger: log.WithField("test", t.Name()),
			File:   f,
		},
		Publish: func(h provider.InlayHint) {
			hints = append(hints, h)
		},
	})
	return hints
}

func providerWithRoutes(routes ...repository.RouteEntry) *Provider {
	p := &Provider{}
	repo := make(repository.RouteRepository)
	for i, r := range routes {
		repo[string(rune('a'+i))] = r
	}
	p.routeCache = repo
	return p
}

func TestResolveInlayHints(t *testing.T) {
	const namespacedController = `<?php
namespace App\Http\Controllers;

class HomeController
{
    public function index() {}
    public function show() {}
}`

	t.Run("fully qualified action matches controller FQN", func(t *testing.T) {
		p := providerWithRoutes(repository.RouteEntry{
			Method: "GET", URI: "/", Action: `App\Http\Controllers\HomeController@index`,
		})
		hints := resolveHints(t, p, namespacedController)
		require.Len(t, hints, 1)
		require.Equal(t, "[GET /]", hints[0].Label)
	})

	t.Run("unqualified class name matches by base name", func(t *testing.T) {
		p := providerWithRoutes(repository.RouteEntry{
			Method: "POST", URI: "home", Action: `HomeController@index`,
		})
		hints := resolveHints(t, p, namespacedController)
		require.Len(t, hints, 1)
		require.Equal(t, "[POST /home]", hints[0].Label)
	})

	t.Run("wrong class name does not match", func(t *testing.T) {
		p := providerWithRoutes(repository.RouteEntry{
			Method: "GET", URI: "/", Action: `OtherController@index`,
		})
		hints := resolveHints(t, p, namespacedController)
		require.Empty(t, hints)
	})

	t.Run("class name that is a suffix of another does not match", func(t *testing.T) {
		p := providerWithRoutes(repository.RouteEntry{
			Method: "GET", URI: "/", Action: `Controller@index`,
		})
		hints := resolveHints(t, p, namespacedController)
		require.Empty(t, hints)
	})

	t.Run("invokable controller with no @ in action", func(t *testing.T) {
		src := `<?php
namespace App\Http\Controllers;

class ShowDashboard
{
    public function __invoke() {}
}`
		p := providerWithRoutes(repository.RouteEntry{
			Method: "GET", URI: "dashboard", Action: `App\Http\Controllers\ShowDashboard`,
		})
		hints := resolveHints(t, p, src)
		require.Len(t, hints, 1)
		require.Equal(t, "[GET /dashboard]", hints[0].Label)
	})

	t.Run("method not found in controller produces no hint", func(t *testing.T) {
		p := providerWithRoutes(repository.RouteEntry{
			Method: "GET", URI: "/", Action: `HomeController@missing`,
		})
		hints := resolveHints(t, p, namespacedController)
		require.Empty(t, hints)
	})

	t.Run("multiple routes produce multiple hints", func(t *testing.T) {
		p := providerWithRoutes(
			repository.RouteEntry{Method: "GET", URI: "/", Action: `HomeController@index`},
			repository.RouteEntry{Method: "GET", URI: "home/{id}", Action: `HomeController@show`},
		)
		hints := resolveHints(t, p, namespacedController)
		require.Len(t, hints, 2)
		labels := []string{hints[0].Label, hints[1].Label}
		require.ElementsMatch(t, []string{"[GET /]", "[GET /home/{id}]"}, labels)
	})

	t.Run("non-class PHP file produces no hints", func(t *testing.T) {
		p := providerWithRoutes(repository.RouteEntry{
			Method: "GET", URI: "/", Action: `HomeController@index`,
		})
		hints := resolveHints(t, p, `<?php echo "hello";`)
		require.Empty(t, hints)
	})

	t.Run("no routes produces no hints", func(t *testing.T) {
		p := &Provider{}
		p.routeCache = make(repository.RouteRepository)
		hints := resolveHints(t, p, namespacedController)
		require.Empty(t, hints)
	})
}

func TestOnFileSaved(t *testing.T) {
	t.Run("file outside routes dir returns nil", func(t *testing.T) {
		p := &Provider{rootPath: "/app"}
		ch := p.OnFileSaved("/app/app/Http/Controllers/HomeController.php")
		require.Nil(t, ch)
	})

	t.Run("nil project returns nil", func(t *testing.T) {
		p := &Provider{rootPath: "/app", project: nil}
		ch := p.OnFileSaved("/app/routes/web.php")
		require.Nil(t, ch)
	})
}
