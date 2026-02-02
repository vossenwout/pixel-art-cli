package cli

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pxcli/internal/client"
)

func TestHistoryCommands_FormatRequests(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantRequest string
	}{
		{
			name:        "undo",
			args:        []string{"undo"},
			wantRequest: "undo",
		},
		{
			name:        "redo",
			args:        []string{"redo"},
			wantRequest: "redo",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubClient{
				response: client.Response{Raw: "ok"},
			}
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
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(stub.requests) != 1 {
				t.Fatalf("expected 1 request, got %d", len(stub.requests))
			}
			if stub.requests[0] != tt.wantRequest {
				t.Fatalf("expected request %q, got %q", tt.wantRequest, stub.requests[0])
			}
			if strings.TrimSpace(buf.String()) != "ok" {
				t.Fatalf("expected ok output, got %q", buf.String())
			}
		})
	}
}

func TestUndoCmd_RevertsCanvasState(t *testing.T) {
	dir := t.TempDir()
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
	daemonCmd.SetArgs([]string{"daemon", "--headless", "--size", "2x2", "--socket", socketPath})

	errCh := make(chan error, 1)
	go func() {
		errCh <- daemonCmd.Execute()
	}()

	waitForPath(t, socketPath)

	cli, err := client.New(socketPath)
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	if _, err := cli.Send("set_pixel 0 0 #ff0000"); err != nil {
		t.Fatalf("unexpected set_pixel error: %v", err)
	}

	buf := &bytes.Buffer{}
	undoCmd := NewRootCmd("dev")
	undoCmd.SetOut(buf)
	undoCmd.SetErr(io.Discard)
	undoCmd.SetArgs([]string{"--socket", socketPath, "undo"})

	if err := undoCmd.Execute(); err != nil {
		t.Fatalf("unexpected undo error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "ok" {
		t.Fatalf("expected ok output, got %q", buf.String())
	}

	getBuf := &bytes.Buffer{}
	getCmd := NewRootCmd("dev")
	getCmd.SetOut(getBuf)
	getCmd.SetErr(io.Discard)
	getCmd.SetArgs([]string{"--socket", socketPath, "get_pixel", "0", "0"})

	if err := getCmd.Execute(); err != nil {
		t.Fatalf("unexpected get_pixel error: %v", err)
	}
	if strings.TrimSpace(getBuf.String()) != "ok #00000000" {
		t.Fatalf("expected ok #00000000 output, got %q", getBuf.String())
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

func TestRedoCmd_NoHistory(t *testing.T) {
	dir := t.TempDir()
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
	daemonCmd.SetArgs([]string{"daemon", "--headless", "--size", "2x2", "--socket", socketPath})

	errCh := make(chan error, 1)
	go func() {
		errCh <- daemonCmd.Execute()
	}()

	waitForPath(t, socketPath)

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--socket", socketPath, "redo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected no_history error")
	}
	if !strings.Contains(err.Error(), "err no_history") {
		t.Fatalf("expected err no_history message, got %q", err.Error())
	}

	cli, err := client.New(socketPath)
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
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
