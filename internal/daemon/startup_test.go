package daemon

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"

	"pxcli/internal/testutil"
)

func TestEnsureDaemonReadyRemovesStalePIDAndSocket(t *testing.T) {
	dir := testutil.TempDir(t)
	pidPath := filepath.Join(dir, "pxcli.pid")
	socketPath := filepath.Join(dir, "pxcli.sock")

	if err := os.WriteFile(pidPath, []byte("1234\n"), 0o644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	if err := os.WriteFile(socketPath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("failed to create socket file: %v", err)
	}

	dialCalled := false
	err := EnsureDaemonReady(pidPath, socketPath, func(pid int) bool {
		if pid != 1234 {
			t.Fatalf("expected pid 1234, got %d", pid)
		}
		return false
	}, func(path string) error {
		dialCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dialCalled {
		t.Fatalf("dialer should not be called when pid file exists")
	}

	assertPathMissing(t, pidPath)
	assertPathMissing(t, socketPath)

	rebind, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("expected to bind after cleanup: %v", err)
	}
	_ = rebind.Close()
}

func TestEnsureDaemonReadyActivePIDReturnsError(t *testing.T) {
	dir := testutil.TempDir(t)
	pidPath := filepath.Join(dir, "pxcli.pid")

	if err := os.WriteFile(pidPath, []byte("5678\n"), 0o644); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	err := EnsureDaemonReady(pidPath, "", func(pid int) bool {
		if pid != 5678 {
			t.Fatalf("expected pid 5678, got %d", pid)
		}
		return true
	}, func(path string) error {
		t.Fatalf("dialer should not be called for active pid")
		return nil
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	var derr Error
	if !errors.As(err, &derr) || derr.Code != "daemon_already_running" {
		t.Fatalf("expected daemon_already_running error, got %v", err)
	}
}

func TestEnsureDaemonReadyRemovesStaleSocket(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")

	if err := os.WriteFile(socketPath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("failed to create socket file: %v", err)
	}

	dialCalled := false
	err := EnsureDaemonReady("", socketPath, nil, func(path string) error {
		dialCalled = true
		return errors.New("no listener")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dialCalled {
		t.Fatalf("expected dialer to be called for existing socket")
	}

	assertPathMissing(t, socketPath)
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %s to be removed", path)
	}
}
