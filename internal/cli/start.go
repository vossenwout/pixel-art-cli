package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"pxcli/internal/config"
)

// NewStartCmd creates the start command skeleton with shared flags.
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the pxcli daemon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("start not implemented")
		},
	}

	cmd.Flags().Bool("headless", config.DefaultHeadless, "Run without a GUI")

	return cmd
}
