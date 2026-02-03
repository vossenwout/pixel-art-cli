package daemon

import "os"

// WindowedOptions provides overrides for the windowed runtime wiring.
type WindowedOptions struct {
	SignalCh <-chan os.Signal
}

func scaledWindowSize(width, height, scale int) (int, int) {
	return width * scale, height * scale
}
