package cli

import "github.com/spf13/cobra"

// NewUndoCmd creates the undo command.
func NewUndoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo the last change",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendCommandRequest(cmd, "undo")
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}

// NewRedoCmd creates the redo command.
func NewRedoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redo",
		Short: "Redo the last undone change",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendCommandRequest(cmd, "redo")
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}
