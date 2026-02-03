//go:build !ebiten

package daemon

import "pxcli/internal/config"

// RunWindowed reports that windowed rendering is unavailable in this build.
func RunWindowed(cfg config.Config, opts WindowedOptions) error {
	return RendererUnavailableError()
}
