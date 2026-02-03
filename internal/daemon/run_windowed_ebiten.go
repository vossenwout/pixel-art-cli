//go:build ebiten

package daemon

import "pxcli/internal/config"

// RunWindowed wires the daemon server with the Ebiten renderer.
func RunWindowed(cfg config.Config, opts WindowedOptions) error {
	return runWindowed(cfg, opts, func(source RenderSource, scale int) (Renderer, error) {
		return NewEbitenRenderer(source, scale)
	})
}
