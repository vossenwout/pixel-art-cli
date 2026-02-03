package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"pxcli/internal/config"
	"pxcli/internal/daemon"
)

var daemonPIDPath = config.DefaultPIDPath

// NewDaemonCmd creates the hidden daemon entrypoint skeleton with shared flags.
func NewDaemonCmd() *cobra.Command {
	var (
		size     string
		scale    int
		headless bool
	)

	cmd := &cobra.Command{
		Use:    "daemon",
		Short:  "Run the pxcli daemon",
		Hidden: true,
		Args:   cobra.NoArgs,
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

			cfg := config.New(
				config.WithSocketPath(socketPath),
				config.WithPIDPath(daemonPIDPath),
				config.WithCanvasSize(width, height),
				config.WithScale(scale),
				config.WithHeadless(headless),
			)

			return daemon.RunHeadless(cfg, daemon.HeadlessOptions{})
		},
	}

	cmd.Flags().StringVar(&size, "size", fmt.Sprintf("%dx%d", config.DefaultCanvasWidth, config.DefaultCanvasHeight), "Canvas size in WxH")
	cmd.Flags().IntVar(&scale, "scale", config.DefaultScale, "Canvas scale (reserved for windowed mode)")
	cmd.Flags().BoolVar(&headless, "headless", config.DefaultHeadless, "Run without a GUI")

	return cmd
}
