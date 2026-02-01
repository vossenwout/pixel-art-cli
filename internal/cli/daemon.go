package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"pxcli/internal/config"
)

// NewDaemonCmd creates the hidden daemon entrypoint skeleton with shared flags.
func NewDaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "daemon",
		Short:  "Run the pxcli daemon",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("daemon not implemented")
		},
	}

	cmd.Flags().Bool("headless", config.DefaultHeadless, "Run without a GUI")

	return cmd
}
