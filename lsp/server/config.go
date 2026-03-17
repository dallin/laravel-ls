package server

// LSPConfig holds configuration passed by the client via initializationOptions
// during the LSP initialize handshake.
//
// e.g. In Neovim (lsp/laravel_ls.lua), set these via init_options:
//
//	init_options = {
//	  inlayHints = {
//	    routes = { enabled = false }
//	  }
//	}
type LSPConfig struct {
	InlayHints InlayHintsConfig `json:"inlayHints"`
}

type InlayHintsConfig struct {
	Routes RouteInlayHintsConfig `json:"routes"`
}

type RouteInlayHintsConfig struct {
	Enabled *bool `json:"enabled"`
}

func (c RouteInlayHintsConfig) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}
