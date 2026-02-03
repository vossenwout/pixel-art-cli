package daemon

import (
	"context"

	"pxcli/internal/canvas"
)

const rendererUnavailableMessage = "windowed mode requires -tags=ebiten"

// Renderer runs the GUI loop and supports external close requests.
type Renderer interface {
	Run(ctx context.Context) error
	RequestClose()
}

// RenderSource provides snapshot and dirty information for rendering.
type RenderSource interface {
	Dirty() bool
	RenderSnapshot() canvas.RenderSnapshot
	Width() int
	Height() int
}

// RendererOptions holds future renderer configuration.
type RendererOptions struct {
	Headless bool
}

// RendererUnavailableError reports missing GUI support.
func RendererUnavailableError() Error {
	return Error{Code: "renderer_unavailable", Message: rendererUnavailableMessage}
}

// ValidateRenderer ensures requested headless/windowed mode is supported.
func ValidateRenderer(headless bool) error {
	if headless {
		return nil
	}
	if RendererAvailable() {
		return nil
	}
	return RendererUnavailableError()
}
