//go:build ebiten

package daemon

// RendererAvailable reports whether GUI rendering is supported in this build.
func RendererAvailable() bool {
	return true
}
