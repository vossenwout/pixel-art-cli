package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pxcli/internal/testutil"
)

func TestStopCmd_StopsDaemonAndRemovesFiles(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")

	restorePID := daemonPIDPath
	daemonPIDPath = pidPath
	t.Cleanup(func() {
		daemonPIDPath = restorePID
	})

	daemonCmd := NewRootCmd("dev")
	daemonCmd.SetOut(io.Discard)
	daemonCmd.SetErr(io.Discard)
	daemonCmd.SetArgs([]string{"daemon", "--headless", "--size", "8x8", "--socket", socketPath})

	errCh := make(chan error, 1)
	go func() {
		errCh <- daemonCmd.Execute()
	}()

	waitForPath(t, socketPath)
	waitForPath(t, pidPath)

	buf := &bytes.Buffer{}
	stopCmd := NewRootCmd("dev")
	stopCmd.SetOut(buf)
	stopCmd.SetErr(buf)
	stopCmd.SetArgs([]string{"stop", "--socket", socketPath})

	if err := stopCmd.Execute(); err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "ok" {
		t.Fatalf("expected ok output, got %q", buf.String())
	}

	if _, err := os.Stat(socketPath); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected socket path removed, got %v", err)
	}
	if _, err := os.Stat(pidPath); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected pid path removed, got %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected daemon error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for daemon to stop")
	}
}

func TestStopCmd_NoDaemon(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"stop", "--socket", socketPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing daemon")
	}
	if !strings.Contains(err.Error(), "err daemon_not_running") {
		t.Fatalf("expected daemon_not_running error, got %q", err.Error())
	}
}
