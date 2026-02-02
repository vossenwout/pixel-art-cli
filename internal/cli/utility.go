package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewGetPixelCmd creates the get_pixel command.
func NewGetPixelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get_pixel <x> <y>",
		Short: "Get a pixel color",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return invalidArgCount(2, len(args))
			}
			if _, err := parseIntArg(args[0], "x"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "y"); err != nil {
				return err
			}
			request := fmt.Sprintf("get_pixel %s %s", args[0], args[1])
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}

// NewExportCmd creates the export command.
func NewExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <filename.png>",
		Short: "Export the canvas to a PNG file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return invalidArgCount(1, len(args))
			}
			absPath, err := filepath.Abs(args[0])
			if err != nil {
				return invalidArgsf("invalid path: %v", err)
			}
			request := fmt.Sprintf("export %s", absPath)
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}
