package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pxcli/internal/client"
)

func TestGetPixelCmd_PrintsColor(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")

	restorePID := daemonPIDPath
	daemonPIDPath = pidPath
	t.Cleanup(func() {
		daemonPIDPath = restorePID
	})
	t.Cleanup(func() {
		if _, err := os.Stat(socketPath); err == nil {
			_, _ = sendRequest(socketPath, "stop\n")
		}
	})

	daemonCmd := NewRootCmd("dev")
	daemonCmd.SetOut(io.Discard)
	daemonCmd.SetErr(io.Discard)
	daemonCmd.SetArgs([]string{"daemon", "--headless", "--size", "4x4", "--socket", socketPath})

	errCh := make(chan error, 1)
	go func() {
		errCh <- daemonCmd.Execute()
	}()

	waitForPath(t, socketPath)

	cli, err := client.New(socketPath)
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	if _, err := cli.Send("set_pixel 0 0 #00ff00"); err != nil {
		t.Fatalf("unexpected set_pixel error: %v", err)
	}

	buf := &bytes.Buffer{}
	getCmd := NewRootCmd("dev")
	getCmd.SetOut(buf)
	getCmd.SetErr(io.Discard)
	getCmd.SetArgs([]string{"--socket", socketPath, "get_pixel", "0", "0"})

	if err := getCmd.Execute(); err != nil {
		t.Fatalf("unexpected get_pixel error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "ok #00ff00ff" {
		t.Fatalf("expected ok #00ff00ff output, got %q", buf.String())
	}

	if _, err := cli.Send("stop"); err != nil {
		t.Fatalf("failed to stop daemon: %v", err)
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

func TestExportCmd_ResolvesAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected getwd error: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("unexpected chdir error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	expected, err := filepath.Abs("out.png")
	if err != nil {
		t.Fatalf("unexpected abs error: %v", err)
	}

	stub := &stubClient{response: client.Response{Raw: "ok"}}
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	buf := &bytes.Buffer{}
	cmd := NewRootCmd("dev")
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"export", "out.png"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected export error: %v", err)
	}
	if len(stub.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(stub.requests))
	}
	if stub.requests[0] != "export "+expected {
		t.Fatalf("expected export request %q, got %q", "export "+expected, stub.requests[0])
	}
	if strings.TrimSpace(buf.String()) != "ok" {
		t.Fatalf("expected ok output, got %q", buf.String())
	}
}

func TestExportCmd_PropagatesIOError(t *testing.T) {
	stub := &stubClient{err: client.Error{Code: "io", Message: "permission denied"}}
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"export", "out.png"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for io failure")
	}
	if !strings.Contains(err.Error(), "err io") {
		t.Fatalf("expected err io message, got %q", err.Error())
	}
}
