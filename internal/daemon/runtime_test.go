package daemon

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pxcli/internal/canvas"
	"pxcli/internal/history"
)

func TestRuntimeStopRequestCleansUp(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")

	grid, err := canvas.New(2, 2)
	if err != nil {
		t.Fatalf("unexpected canvas error: %v", err)
	}
	manager := history.New(grid)
	stopper := NewStopper()
	server, err := NewServer(socketPath, NewHandler(manager, stopper.Stop))
	if err != nil {
		t.Fatalf("unexpected server error: %v", err)
	}
	if err := WritePID(pidPath, os.Getpid()); err != nil {
		t.Fatalf("unexpected pid write error: %v", err)
	}

	runtime, err := NewRuntime(server, RuntimeOptions{
		PIDPath:    pidPath,
		SocketPath: socketPath,
		StopCh:     stopper.Done(),
	})
	if err != nil {
		t.Fatalf("unexpected runtime error: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		done <- runtime.Run()
	}()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("unexpected dial error: %v", err)
	}
	if _, err := io.WriteString(conn, "stop\n"); err != nil {
		_ = conn.Close()
		t.Fatalf("unexpected write error: %v", err)
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		_ = conn.Close()
		t.Fatalf("unexpected read error: %v", err)
	}
	_ = conn.Close()
	if line != "ok\n" {
		t.Fatalf("expected ok response, got %q", line)
	}

	assertRuntimeDone(t, done)
	assertPathMissing(t, pidPath)
	assertPathMissing(t, socketPath)
}

func TestRuntimeSignalShutdownWithoutClients(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")

	server, err := NewServer(socketPath, stubHandler{response: "ok"})
	if err != nil {
		t.Fatalf("unexpected server error: %v", err)
	}
	if err := WritePID(pidPath, os.Getpid()); err != nil {
		t.Fatalf("unexpected pid write error: %v", err)
	}

	signalCh := make(chan os.Signal, 1)
	runtime, err := NewRuntime(server, RuntimeOptions{
		PIDPath:    pidPath,
		SocketPath: socketPath,
		SignalCh:   signalCh,
	})
	if err != nil {
		t.Fatalf("unexpected runtime error: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		done <- runtime.Run()
	}()

	waitForPath(t, socketPath)
	signalCh <- os.Interrupt

	assertRuntimeDone(t, done)
	assertPathMissing(t, pidPath)
	assertPathMissing(t, socketPath)
}

func assertRuntimeDone(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected runtime error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("runtime did not shut down in time")
	}
}

func waitForPath(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, err := os.Stat(path)
		if err == nil {
			return
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("unexpected stat error for %s: %v", path, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := os.Stat(path); err != nil && errors.Is(err, os.ErrNotExist) {
		t.Fatalf("timed out waiting for %s", path)
	}
}
