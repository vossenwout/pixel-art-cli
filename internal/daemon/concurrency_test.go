package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"pxcli/internal/config"
)

func TestHeadlessRuntimeConcurrentRequests(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")
	cfg := config.New(
		config.WithSocketPath(socketPath),
		config.WithPIDPath(pidPath),
		config.WithCanvasSize(10, 10),
	)

	done := startHeadlessRuntime(t, cfg)
	t.Cleanup(func() {
		if _, err := os.Stat(socketPath); err == nil {
			_, _ = sendRequest(socketPath, "stop\n")
		}
	})

	type coord struct {
		x int
		y int
	}
	coords := make([]coord, 0, 25)
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			coords = append(coords, coord{x: x, y: y})
		}
	}

	errs := make(chan error, len(coords))
	var wg sync.WaitGroup
	for _, c := range coords {
		c := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := sendRequest(socketPath, fmt.Sprintf("set_pixel %d %d #00ff00\n", c.x, c.y))
			if err != nil {
				errs <- err
				return
			}
			if response != "ok\n" {
				errs <- fmt.Errorf("unexpected response for (%d,%d): %q", c.x, c.y, response)
			}
		}()
	}

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for concurrent requests")
	}

	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent request failed: %v", err)
		}
	}

	for _, c := range coords {
		response, err := sendRequest(socketPath, fmt.Sprintf("get_pixel %d %d\n", c.x, c.y))
		if err != nil {
			t.Fatalf("unexpected get_pixel error: %v", err)
		}
		if response != "ok #00ff00ff\n" {
			t.Fatalf("expected green pixel at (%d,%d), got %q", c.x, c.y, response)
		}
	}

	response := mustSendRequest(t, socketPath, "stop\n")
	if response != "ok\n" {
		t.Fatalf("expected ok stop response, got %q", response)
	}

	assertRuntimeDone(t, done)
	assertPathMissing(t, pidPath)
	assertPathMissing(t, socketPath)
}
