package daemon

import (
	"bufio"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pxcli/internal/config"
	"pxcli/internal/testutil"
)

func TestHeadlessRuntimeHandlesCommands(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")
	cfg := config.New(
		config.WithSocketPath(socketPath),
		config.WithPIDPath(pidPath),
		config.WithCanvasSize(2, 2),
	)

	done := startHeadlessRuntime(t, cfg)
	t.Cleanup(func() {
		if _, err := os.Stat(socketPath); err == nil {
			_, _ = sendRequest(socketPath, "stop\n")
		}
	})

	response := mustSendRequest(t, socketPath, "set_pixel 0 0 #ff0000\n")
	if response != "ok\n" {
		t.Fatalf("expected ok response, got %q", response)
	}

	response = mustSendRequest(t, socketPath, "get_pixel 0 0\n")
	if response != "ok #ff0000ff\n" {
		t.Fatalf("expected red pixel, got %q", response)
	}

	response = mustSendRequest(t, socketPath, "stop\n")
	if response != "ok\n" {
		t.Fatalf("expected ok stop response, got %q", response)
	}

	assertRuntimeDone(t, done)
	assertPathMissing(t, pidPath)
	assertPathMissing(t, socketPath)
}

func TestHeadlessRuntimeWithoutDisplay(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")
	cfg := config.New(
		config.WithSocketPath(socketPath),
		config.WithPIDPath(pidPath),
		config.WithCanvasSize(1, 1),
	)

	previousDisplay, hadDisplay := os.LookupEnv("DISPLAY")
	if err := os.Unsetenv("DISPLAY"); err != nil {
		t.Fatalf("unexpected DISPLAY unset error: %v", err)
	}
	t.Cleanup(func() {
		if hadDisplay {
			_ = os.Setenv("DISPLAY", previousDisplay)
		} else {
			_ = os.Unsetenv("DISPLAY")
		}
		if _, err := os.Stat(socketPath); err == nil {
			_, _ = sendRequest(socketPath, "stop\n")
		}
	})

	done := startHeadlessRuntime(t, cfg)

	response := mustSendRequest(t, socketPath, "clear\n")
	if response != "ok\n" {
		t.Fatalf("expected ok response, got %q", response)
	}

	response = mustSendRequest(t, socketPath, "stop\n")
	if response != "ok\n" {
		t.Fatalf("expected ok stop response, got %q", response)
	}

	assertRuntimeDone(t, done)
}

func startHeadlessRuntime(t *testing.T, cfg config.Config) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)
	go func() {
		done <- RunHeadless(cfg, HeadlessOptions{SignalCh: signalCh})
	}()
	waitForPath(t, cfg.SocketPath)
	return done
}

func mustSendRequest(t *testing.T, socketPath, request string) string {
	t.Helper()
	line, err := sendRequest(socketPath, request)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	return line
}

func sendRequest(socketPath, request string) (string, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if !strings.HasSuffix(request, "\n") {
		request += "\n"
	}
	if _, err := io.WriteString(conn, request); err != nil {
		return "", err
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}
