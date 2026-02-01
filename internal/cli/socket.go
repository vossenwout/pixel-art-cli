package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// SocketPath returns the resolved socket path for a command.
func SocketPath(cmd *cobra.Command) (string, error) {
	socketPath, err := cmd.Flags().GetString("socket")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(socketPath) == "" {
		return "", fmt.Errorf("socket path must not be empty")
	}
	return socketPath, nil
}
