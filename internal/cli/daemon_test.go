package cli

import (
	"bufio"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseCanvasSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		input       string
		wantW       int
		wantH       int
		wantErr     bool
		errContains string
	}{
		{
			name:  "valid size",
			input: "32x16",
			wantW: 32,
			wantH: 16,
		},
		{
			name:        "missing height",
			input:       "10",
			wantErr:     true,
			errContains: "expected WxH",
		},
		{
			name:        "non-positive width",
			input:       "0x10",
			wantErr:     true,
			errContains: "width and height must be positive",
		},
		{
			name:        "non-positive height",
			input:       "10x0",
			wantErr:     true,
			errContains: "width and height must be positive",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotW, gotH, err := parseCanvasSize(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tt.input)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Fatalf("expected %dx%d, got %dx%d", tt.wantW, tt.wantH, gotW, gotH)
			}
		})
	}
}

func TestDaemonCmd_StartsHeadlessRuntime(t *testing.T) {
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

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"daemon", "--headless", "--size", "32x32", "--socket", socketPath})

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Execute()
	}()

	waitForPath(t, socketPath)

	response := mustSendRequest(t, socketPath, "clear\n")
	if response != "ok\n" {
		t.Fatalf("expected ok response, got %q", response)
	}

	response = mustSendRequest(t, socketPath, "stop\n")
	if response != "ok\n" {
		t.Fatalf("expected ok stop response, got %q", response)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected daemon command error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for daemon command to exit")
	}
}

func TestDaemonCmd_InvalidSize(t *testing.T) {
	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"daemon", "--size", "10"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid size")
	}
	if !strings.Contains(err.Error(), "invalid size") {
		t.Fatalf("expected size validation error, got %q", err.Error())
	}
}

func TestDaemonCmd_InvalidScale(t *testing.T) {
	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"daemon", "--scale", "0"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid scale")
	}
	if !strings.Contains(err.Error(), "invalid scale") {
		t.Fatalf("expected scale validation error, got %q", err.Error())
	}
}

func waitForPath(t *testing.T, path string) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %s", path)
		case <-ticker.C:
			if _, err := os.Stat(path); err == nil {
				return
			}
		}
	}
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
	if err := conn.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return "", err
	}
	if _, err := io.WriteString(conn, request); err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)
	return reader.ReadString('\n')
}
