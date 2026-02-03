//go:build ebiten

package daemon

import (
	"context"
	"os"

	"pxcli/internal/canvas"
	"pxcli/internal/config"
	"pxcli/internal/history"
)

// RunWindowed wires the daemon server with the Ebiten renderer.
func RunWindowed(cfg config.Config, opts WindowedOptions) error {
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

	renderer, err := NewEbitenRenderer(grid, cfg.Scale)
	if err != nil {
		return err
	}

	handler := NewHandler(manager, func() {
		stopper.Stop()
		renderer.RequestClose()
	})

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtimeErrCh := make(chan error, 1)
	go func() {
		runtimeErrCh <- runtime.Run()
		cancel()
	}()

	renderErr := renderer.Run(ctx)
	stopper.Stop()
	runtimeErr := <-runtimeErrCh

	if renderErr != nil {
		return renderErr
	}
	if runtimeErr != nil {
		return runtimeErr
	}
	return nil
}
