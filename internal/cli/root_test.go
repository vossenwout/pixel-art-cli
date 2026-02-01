package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_UnknownCommandPrintsUsage(t *testing.T) {
	cmd := NewRootCmd("dev")
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"does-not-exist"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for unknown command")
	}

	output := buf.String()
	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected usage in output, got: %q", output)
	}
}

func TestRootCmd_VersionFlag(t *testing.T) {
	cmd := NewRootCmd("1.2.3")
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "1.2.3" {
		t.Fatalf("expected version output %q, got %q", "1.2.3", output)
	}
}
