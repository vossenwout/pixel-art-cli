package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"pxcli/internal/config"
)

func TestSocketPath_DefaultForClientCommands(t *testing.T) {
	cmd := NewRootCmd("dev")
	var got string
	clientCmd := &cobra.Command{
		Use: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			got, err = SocketPath(cmd)
			return err
		},
	}
	cmd.AddCommand(clientCmd)
	cmd.SetArgs([]string{"client"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != config.DefaultSocketPath {
		t.Fatalf("expected default socket %q, got %q", config.DefaultSocketPath, got)
	}
}

func TestSocketPath_EmptyReturnsError(t *testing.T) {
	cmd := NewRootCmd("dev")
	clientCmd := &cobra.Command{
		Use: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := SocketPath(cmd)
			return err
		},
	}
	cmd.AddCommand(clientCmd)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"client", "--socket", ""})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for empty socket path")
	}
	if !strings.Contains(err.Error(), "socket path must not be empty") {
		t.Fatalf("expected socket validation error, got %q", err.Error())
	}
}

func TestStartCmd_HeadlessFlagDefault(t *testing.T) {
	cmd := NewRootCmd("dev")
	startCmd, _, err := cmd.Find([]string{"start"})
	if err != nil {
		t.Fatalf("unexpected error finding start command: %v", err)
	}
	flag := startCmd.Flags().Lookup("headless")
	if flag == nil {
		t.Fatalf("expected headless flag on start command")
	}
	if flag.DefValue != "false" {
		t.Fatalf("expected headless default false, got %q", flag.DefValue)
	}
}

func TestDaemonCmd_HeadlessFlagDefault(t *testing.T) {
	cmd := NewRootCmd("dev")
	daemonCmd, _, err := cmd.Find([]string{"daemon"})
	if err != nil {
		t.Fatalf("unexpected error finding daemon command: %v", err)
	}
	flag := daemonCmd.Flags().Lookup("headless")
	if flag == nil {
		t.Fatalf("expected headless flag on daemon command")
	}
	if flag.DefValue != "false" {
		t.Fatalf("expected headless default false, got %q", flag.DefValue)
	}
}
