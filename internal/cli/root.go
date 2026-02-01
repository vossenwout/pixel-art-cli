package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd returns the root pxcli command.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pxcli",
		Short:         "pxcli is a pixel art CLI",
		SilenceUsage:  false,
		SilenceErrors: false,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.Version = version
	cmd.SetVersionTemplate("{{.Version}}\n")

	return cmd
}
