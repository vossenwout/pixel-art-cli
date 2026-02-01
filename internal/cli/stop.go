package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"pxcli/internal/client"
)

var (
	stopNewClient       = client.New
	stopWaitTimeout     = 2 * time.Second
	stopPollInterval    = 20 * time.Millisecond
	stopWaitForShutdown = waitForShutdown
)

// NewStopCmd creates the stop command.
func NewStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the pxcli daemon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			socketPath, err := SocketPath(cmd)
			if err != nil {
				return err
			}
			cli, err := stopNewClient(socketPath)
			if err != nil {
				return err
			}
			resp, err := cli.Send("stop")
			if err != nil {
				return formatClientError(err)
			}
			if err := stopWaitForShutdown(socketPath, daemonPIDPath, stopWaitTimeout); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), resp.Raw)
			return nil
		},
	}

	return cmd
}

func formatClientError(err error) error {
	if err == nil {
		return nil
	}
	var clientErr client.Error
	if errors.As(err, &clientErr) {
		if strings.TrimSpace(clientErr.Message) == "" {
			return fmt.Errorf("err %s", clientErr.Code)
		}
		return fmt.Errorf("err %s %s", clientErr.Code, clientErr.Message)
	}
	return err
}

func waitForShutdown(socketPath, pidPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		socketGone, err := pathGone(socketPath)
		if err != nil {
			return err
		}
		pidGone, err := pathGone(pidPath)
		if err != nil {
			return err
		}
		if socketGone && pidGone {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for daemon to stop")
		}
		time.Sleep(stopPollInterval)
	}
}

func pathGone(path string) (bool, error) {
	if strings.TrimSpace(path) == "" {
		return true, nil
	}
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	return false, err
}
