package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"pxcli/internal/client"
	"pxcli/internal/testutil"
)

func TestBuildDaemonArgs(t *testing.T) {
	got := buildDaemonArgs("/tmp/pxcli.sock", "8x8", 12, false)
	want := []string{
		"daemon",
		"--size", "8x8",
		"--scale", "12",
		"--headless=false",
		"--socket", "/tmp/pxcli.sock",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected args %v, got %v", want, got)
	}
}

func TestStartCmd_InvalidScale(t *testing.T) {
	cases := []string{"0", "-1"}
	for _, scale := range cases {
		cmd := NewRootCmd("dev")
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs([]string{"start", "--scale", scale})

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected error for invalid scale %s", scale)
		}
		if !strings.Contains(err.Error(), "invalid scale") {
			t.Fatalf("expected scale validation error for %s, got %q", scale, err.Error())
		}
	}
}

func TestStartCmd_StartsDaemonAndDetectsRunning(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")

	restorePID := daemonPIDPath
	daemonPIDPath = pidPath
	t.Cleanup(func() {
		daemonPIDPath = restorePID
	})

	restoreSpawn := startSpawnDaemon
	startSpawnDaemon = func(binary string, args []string) (daemonProcess, error) {
		cmd := NewRootCmd("dev")
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)

		go func() {
			_ = cmd.Execute()
		}()

		return daemonProcess{pid: os.Getpid(), release: func() error { return nil }}, nil
	}
	t.Cleanup(func() {
		startSpawnDaemon = restoreSpawn
	})

	buf := &bytes.Buffer{}
	cmd := NewRootCmd("dev")
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"start", "--size", "8x8", "--headless", "--socket", socketPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}

	gotPID := strings.TrimSpace(buf.String())
	if gotPID == "" {
		t.Fatalf("expected daemon PID output")
	}
	if _, err := strconv.Atoi(gotPID); err != nil {
		t.Fatalf("expected numeric PID output, got %q", gotPID)
	}

	cli, err := client.New(socketPath)
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	resp, err := cli.Send("get_pixel 7 7")
	if err != nil {
		t.Fatalf("expected ok response, got %v", err)
	}
	if resp.Payload != "#00000000" {
		t.Fatalf("expected transparent pixel, got %q", resp.Payload)
	}

	_, err = cli.Send("get_pixel 8 8")
	if err == nil {
		t.Fatalf("expected out_of_bounds error")
	}
	var clientErr client.Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected client error, got %v", err)
	}
	if clientErr.Code != "out_of_bounds" {
		t.Fatalf("expected out_of_bounds error, got %q", clientErr.Code)
	}

	cmd2 := NewRootCmd("dev")
	cmd2.SetOut(io.Discard)
	cmd2.SetErr(io.Discard)
	cmd2.SetArgs([]string{"start", "--socket", socketPath})

	err = cmd2.Execute()
	if err == nil {
		t.Fatalf("expected daemon_already_running error")
	}
	if !strings.Contains(err.Error(), "err daemon_already_running") {
		t.Fatalf("expected daemon_already_running error, got %q", err.Error())
	}

	if _, err := cli.Send("stop"); err != nil {
		t.Fatalf("failed to stop daemon: %v", err)
	}
}
