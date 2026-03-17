package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool { return &b }

func TestRouteInlayHintsIsEnabled(t *testing.T) {
	t.Run("nil pointer defaults to enabled", func(t *testing.T) {
		c := RouteInlayHintsConfig{Enabled: nil}
		require.True(t, c.IsEnabled())
	})

	t.Run("explicit true is enabled", func(t *testing.T) {
		c := RouteInlayHintsConfig{Enabled: boolPtr(true)}
		require.True(t, c.IsEnabled())
	})

	t.Run("explicit false is disabled", func(t *testing.T) {
		c := RouteInlayHintsConfig{Enabled: boolPtr(false)}
		require.False(t, c.IsEnabled())
	})
}
