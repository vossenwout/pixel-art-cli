package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"pxcli/internal/client"
)

type requestSender interface {
	Send(request string) (client.Response, error)
}

type clientFactory func(socketPath string) (requestSender, error)

var drawNewClient clientFactory = func(socketPath string) (requestSender, error) {
	return client.New(socketPath)
}

// NewSetPixelCmd creates the set_pixel command.
func NewSetPixelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_pixel <x> <y> <color>",
		Short: "Set a pixel color",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return invalidArgCount(3, len(args))
			}
			if _, err := parseIntArg(args[0], "x"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "y"); err != nil {
				return err
			}
			request := fmt.Sprintf("set_pixel %s %s %s", args[0], args[1], args[2])
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}

// NewFillRectCmd creates the fill_rect command.
func NewFillRectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill_rect <x> <y> <w> <h> <color>",
		Short: "Fill a rectangle on the canvas",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 5 {
				return invalidArgCount(5, len(args))
			}
			if _, err := parseIntArg(args[0], "x"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "y"); err != nil {
				return err
			}
			w, err := parseIntArg(args[2], "w")
			if err != nil {
				return err
			}
			h, err := parseIntArg(args[3], "h")
			if err != nil {
				return err
			}
			if w <= 0 {
				return invalidArgsf("w must be > 0")
			}
			if h <= 0 {
				return invalidArgsf("h must be > 0")
			}
			request := fmt.Sprintf("fill_rect %s %s %s %s %s", args[0], args[1], args[2], args[3], args[4])
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}

// NewLineCmd creates the line command.
func NewLineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "line <x1> <y1> <x2> <y2> <color>",
		Short: "Draw a line on the canvas",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 5 {
				return invalidArgCount(5, len(args))
			}
			if _, err := parseIntArg(args[0], "x1"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[1], "y1"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[2], "x2"); err != nil {
				return err
			}
			if _, err := parseIntArg(args[3], "y2"); err != nil {
				return err
			}
			request := fmt.Sprintf("line %s %s %s %s %s", args[0], args[1], args[2], args[3], args[4])
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}

// NewClearCmd creates the clear command.
func NewClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear [color]",
		Short: "Clear the canvas",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return invalidArgCount(1, len(args))
			}
			request := "clear"
			if len(args) == 1 {
				request = fmt.Sprintf("clear %s", args[0])
			}
			return sendCommandRequest(cmd, request)
		},
	}
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func sendCommandRequest(cmd *cobra.Command, request string) error {
	socketPath, err := SocketPath(cmd)
	if err != nil {
		return err
	}
	cli, err := drawNewClient(socketPath)
	if err != nil {
		return err
	}
	resp, err := cli.Send(request)
	if err != nil {
		return formatClientError(err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), resp.Raw)
	return nil
}

func invalidArgCount(expected, got int) error {
	return invalidArgsf("expected %d args, got %d", expected, got)
}

func parseIntArg(value, name string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, invalidArgsf("%s must be an integer", name)
	}
	return parsed, nil
}

func invalidArgsf(format string, args ...any) error {
	message := strings.TrimSpace(fmt.Sprintf(format, args...))
	if message == "" {
		return fmt.Errorf("err invalid_args")
	}
	return fmt.Errorf("err invalid_args %s", message)
}
