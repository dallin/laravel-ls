package route

import (
	"testing"

	"github.com/laravel-ls/laravel-ls/file"
	"github.com/laravel-ls/laravel-ls/parser"
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
