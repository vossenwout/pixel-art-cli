package cli

import (
	"github.com/spf13/cobra"

	"pxcli/internal/config"
)

// NewRootCmd returns the root pxcli command.
func NewRootCmd(version string) *cobra.Command {
	var socketPath string

	cmd := &cobra.Command{
		Use:           "pxcli",
		Short:         "pxcli is a pixel art CLI",
		SilenceUsage:  false,
		SilenceErrors: false,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			_, err := SocketPath(cmd)
			return err
		},
	}

	cmd.PersistentFlags().StringVar(&socketPath, "socket", config.DefaultSocketPath, "Unix socket path")

	cmd.Version = version
	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.AddCommand(NewStartCmd())
	cmd.AddCommand(NewDaemonCmd())
	cmd.AddCommand(NewStopCmd())
	cmd.AddCommand(NewSetPixelCmd())
	cmd.AddCommand(NewFillRectCmd())
	cmd.AddCommand(NewLineCmd())
	cmd.AddCommand(NewClearCmd())
	cmd.AddCommand(NewGetPixelCmd())
	cmd.AddCommand(NewExportCmd())

	return cmd
}
