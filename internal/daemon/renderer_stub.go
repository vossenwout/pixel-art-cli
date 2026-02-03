//go:build !ebiten

package daemon

import "context"

// StubRenderer is a no-op renderer used when GUI support is unavailable.
type StubRenderer struct{}

func (r *StubRenderer) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (r *StubRenderer) RequestClose() {}

// RendererAvailable reports whether GUI rendering is supported in this build.
func RendererAvailable() bool {
	return false
}
