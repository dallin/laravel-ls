package provider

// FileSaveProvider can be implemented by providers that need to react to
// file save events (e.g. to invalidate caches).
// OnFileSaved returns a channel that closes when the provider's cache has
// been re-warmed, or nil if no async work is needed (or the file was not
// relevant to this provider).
type FileSaveProvider interface {
	OnFileSaved(filename string) <-chan struct{}
}
