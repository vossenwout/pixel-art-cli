package daemon

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"pxcli/internal/client"
	"pxcli/internal/config"
	"pxcli/internal/testutil"
)

func TestScaledWindowSize(t *testing.T) {
	t.Parallel()
	gotW, gotH := scaledWindowSize(8, 8, 10)
	if gotW != 80 || gotH != 80 {
		t.Fatalf("expected 80x80, got %dx%d", gotW, gotH)
	}
}

func TestWindowedStopRequestsRendererClose(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")

	renderer := newBlockingRenderer(true)
	cfg := config.New(
		config.WithSocketPath(socketPath),
		config.WithPIDPath(pidPath),
		config.WithCanvasSize(2, 2),
		config.WithScale(2),
	)

	done := make(chan error, 1)
	go func() {
		done <- runWindowed(cfg, WindowedOptions{}, renderer.Factory())
	}()

	waitForPath(t, socketPath)

	cli, err := client.New(socketPath)
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	resp, err := cli.Send("stop")
	if err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}
	if resp.Raw != "ok" {
		t.Fatalf("expected ok response, got %q", resp.Raw)
	}

	select {
	case <-renderer.requestCloseCh:
	case <-time.After(time.Second):
		t.Fatalf("expected renderer close request")
	}

	select {
	case err := <-done:
		t.Fatalf("expected windowed runtime to wait for renderer close, got %v", err)
	default:
	}

	renderer.CloseWindow()
	assertRuntimeDone(t, done)
	assertPathMissing(t, pidPath)
	assertPathMissing(t, socketPath)
}

func TestWindowedCloseTriggersShutdown(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")

	renderer := newBlockingRenderer(false)
	cfg := config.New(
		config.WithSocketPath(socketPath),
		config.WithPIDPath(pidPath),
		config.WithCanvasSize(2, 2),
		config.WithScale(2),
	)

	done := make(chan error, 1)
	go func() {
		done <- runWindowed(cfg, WindowedOptions{}, renderer.Factory())
	}()

	waitForPath(t, socketPath)
	renderer.CloseWindow()

	assertRuntimeDone(t, done)
	assertPathMissing(t, pidPath)
	assertPathMissing(t, socketPath)
}

type blockingRenderer struct {
	ignoreCtx      bool
	requestCloseCh chan struct{}
	runExitCh      chan struct{}
	runStartedCh   chan struct{}
	requestOnce    sync.Once
	closeOnce      sync.Once
}

func newBlockingRenderer(ignoreCtx bool) *blockingRenderer {
	return &blockingRenderer{
		ignoreCtx:      ignoreCtx,
		requestCloseCh: make(chan struct{}),
		runExitCh:      make(chan struct{}),
		runStartedCh:   make(chan struct{}),
	}
}

func (r *blockingRenderer) Run(ctx context.Context) error {
	close(r.runStartedCh)
	if r.ignoreCtx || ctx == nil {
		<-r.runExitCh
		return nil
	}
	select {
	case <-r.runExitCh:
		return nil
	case <-ctx.Done():
		return nil
	}
}

func (r *blockingRenderer) RequestClose() {
	r.requestOnce.Do(func() {
		close(r.requestCloseCh)
	})
}

func (r *blockingRenderer) CloseWindow() {
	r.closeOnce.Do(func() {
		close(r.runExitCh)
	})
}

func (r *blockingRenderer) Factory() rendererFactory {
	return func(source RenderSource, scale int) (Renderer, error) {
		return r, nil
	}
}
