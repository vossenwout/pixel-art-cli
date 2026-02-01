package daemon

import (
	"os"

	"pxcli/internal/canvas"
	"pxcli/internal/config"
	"pxcli/internal/history"
)

// HeadlessOptions provides overrides for headless runtime wiring.
type HeadlessOptions struct {
	SignalCh <-chan os.Signal
}

// RunHeadless wires the daemon server and blocks until shutdown.
func RunHeadless(cfg config.Config, opts HeadlessOptions) error {
	socketPath := cfg.SocketPath
	pidPath := cfg.PIDPath

	if err := EnsureDaemonReady(pidPath, socketPath, nil, nil); err != nil {
		return err
	}

	grid, err := canvas.New(cfg.CanvasWidth, cfg.CanvasHeight)
	if err != nil {
		return err
	}
	manager := history.New(grid)
	stopper := NewStopper()
	handler := NewHandler(manager, stopper.Stop)

	server, err := NewServer(socketPath, handler)
	if err != nil {
		return err
	}
	if err := WritePID(pidPath, os.Getpid()); err != nil {
		_ = server.Close()
		_ = CleanupFiles(pidPath, socketPath)
		return err
	}

	runtime, err := NewRuntime(server, RuntimeOptions{
		PIDPath:    pidPath,
		SocketPath: socketPath,
		StopCh:     stopper.Done(),
		SignalCh:   opts.SignalCh,
	})
	if err != nil {
		_ = server.Close()
		_ = CleanupFiles(pidPath, socketPath)
		return err
	}

	return runtime.Run()
}
