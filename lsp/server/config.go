package server

// LSPConfig holds configuration passed by the client via initializationOptions
// during the LSP initialize handshake.
//
// In Neovim (lsp/laravel_ls.lua), set these via init_options:
//
//	init_options = {
//	  capabilities = {
//	    inlayHints = {
//	      routes = { enabled = false }
//	    }
//	  }
//	}
type LSPConfig struct {
	Capabilities CapabilitiesConfig `json:"capabilities"`
}

// CapabilitiesConfig controls which server capabilities are active.
type CapabilitiesConfig struct {
	InlayHints InlayHintsConfig `json:"inlayHints"`
}

// InlayHintsConfig controls inlay hint support per feature area.
type InlayHintsConfig struct {
	Routes RouteInlayHintsConfig `json:"routes"`
}

// RouteInlayHintsConfig controls route inlay hints, which display the HTTP
// method and URI above matching controller action methods.
// Enabled defaults to true when not explicitly set.
type RouteInlayHintsConfig struct {
	Enabled *bool `json:"enabled"`
}

// IsEnabled returns true unless explicitly set to false.
func (c RouteInlayHintsConfig) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}
