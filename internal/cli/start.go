package cli

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"pxcli/internal/config"
	"pxcli/internal/daemon"
)

type daemonProcess struct {
	pid     int
	release func() error
}

var (
	startSpawnDaemon  = spawnDaemonProcess
	startEnsureReady  = ensureDaemonReady
	startWaitForReady = waitForSocketReady
	startWaitTimeout  = 2 * time.Second
	startDialTimeout  = 200 * time.Millisecond
	startPollInterval = 20 * time.Millisecond
)

// NewStartCmd creates the start command with shared flags.
func NewStartCmd() *cobra.Command {
	var (
		size     string
		scale    int
		headless bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the pxcli daemon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			socketPath, err := SocketPath(cmd)
			if err != nil {
				return err
			}
			width, height, err := parseCanvasSize(size)
			if err != nil {
				return err
			}
			if scale <= 0 {
				return fmt.Errorf("invalid scale %d: must be > 0", scale)
			}
			if err := daemon.ValidateRenderer(headless); err != nil {
				return formatDaemonError(err)
			}
			if err := startEnsureReady(daemonPIDPath, socketPath); err != nil {
				return formatDaemonError(err)
			}

			daemonArgs := buildDaemonArgs(socketPath, fmt.Sprintf("%dx%d", width, height), scale, headless)
			executable, err := os.Executable()
			if err != nil {
				return err
			}

			proc, err := startSpawnDaemon(executable, daemonArgs)
			if err != nil {
				return err
			}
			if proc.release != nil {
				defer proc.release()
			}

			if err := startWaitForReady(socketPath, startWaitTimeout); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), proc.pid)
			return nil
		},
	}

	cmd.Flags().StringVar(&size, "size", fmt.Sprintf("%dx%d", config.DefaultCanvasWidth, config.DefaultCanvasHeight), "Canvas size in WxH")
	cmd.Flags().IntVar(&scale, "scale", config.DefaultScale, "Canvas scale (reserved for windowed mode)")
	cmd.Flags().BoolVar(&headless, "headless", config.DefaultHeadless, "Run without a GUI")

	return cmd
}

func buildDaemonArgs(socketPath, size string, scale int, headless bool) []string {
	args := []string{
		"daemon",
		"--size", size,
		"--scale", strconv.Itoa(scale),
		fmt.Sprintf("--headless=%t", headless),
	}
	if strings.TrimSpace(socketPath) != "" {
		args = append(args, "--socket", socketPath)
	}
	return args
}

func spawnDaemonProcess(binary string, args []string) (daemonProcess, error) {
	cmd := exec.Command(binary, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return daemonProcess{}, err
	}
	release := func() error {
		if cmd.Process == nil {
			return nil
		}
		return cmd.Process.Release()
	}
	return daemonProcess{pid: cmd.Process.Pid, release: release}, nil
}

func ensureDaemonReady(pidPath, socketPath string) error {
	return daemon.EnsureDaemonReady(pidPath, socketPath, nil, nil)
}

func waitForSocketReady(socketPath string, timeout time.Duration) error {
	if strings.TrimSpace(socketPath) == "" {
		return fmt.Errorf("socket path must not be empty")
	}

	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("unix", socketPath, startDialTimeout)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for daemon to start")
		}
		time.Sleep(startPollInterval)
	}
}
