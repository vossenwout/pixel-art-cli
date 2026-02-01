package daemon

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Error represents daemon lifecycle errors.
type Error struct {
	Code    string
	Message string
}

func (e Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

// LivenessFunc returns true when the provided PID is alive.
type LivenessFunc func(pid int) bool

// DialFunc attempts a connection to the daemon socket.
type DialFunc func(socketPath string) error

const defaultDialTimeout = 200 * time.Millisecond

var errInvalidPID = errors.New("invalid pid file contents")

// EnsureDaemonReady removes stale PID/socket files and reports active daemons.
func EnsureDaemonReady(pidPath, socketPath string, isAlive LivenessFunc, dial DialFunc) error {
	if isAlive == nil {
		isAlive = processAlive
	}
	if dial == nil {
		dial = dialSocket
	}

	if pidPath != "" {
		pid, err := readPID(pidPath)
		if err == nil {
			if isAlive(pid) {
				return Error{Code: "daemon_already_running", Message: fmt.Sprintf("pid %d is still running", pid)}
			}
			if err := removeIfExists(pidPath); err != nil {
				return err
			}
			if socketPath != "" {
				if err := removeIfExists(socketPath); err != nil {
					return err
				}
			}
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, errInvalidPID) {
			return err
		}
		if errors.Is(err, errInvalidPID) {
			if err := removeIfExists(pidPath); err != nil {
				return err
			}
			if socketPath != "" {
				if err := removeIfExists(socketPath); err != nil {
					return err
				}
			}
			return nil
		}
	}

	if socketPath == "" {
		return nil
	}
	if _, err := os.Stat(socketPath); err == nil {
		if err := dial(socketPath); err == nil {
			return Error{Code: "daemon_already_running", Message: "socket is active"}
		}
		if err := removeIfExists(socketPath); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

// WritePID persists the daemon PID to disk.
func WritePID(pidPath string, pid int) error {
	if pidPath == "" {
		return nil
	}
	if pid <= 0 {
		return Error{Code: "invalid_pid", Message: "pid must be positive"}
	}
	payload := []byte(strconv.Itoa(pid) + "\n")
	return os.WriteFile(pidPath, payload, 0o644)
}

// CleanupFiles removes PID and socket files when the daemon exits.
func CleanupFiles(pidPath, socketPath string) error {
	if err := removeIfExists(socketPath); err != nil {
		return err
	}
	if err := removeIfExists(pidPath); err != nil {
		return err
	}
	return nil
}

func readPID(pidPath string) (int, error) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	raw := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(raw)
	if err != nil || pid <= 0 {
		return 0, errInvalidPID
	}
	return pid, nil
}

func removeIfExists(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}

func dialSocket(socketPath string) error {
	conn, err := net.DialTimeout("unix", socketPath, defaultDialTimeout)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
