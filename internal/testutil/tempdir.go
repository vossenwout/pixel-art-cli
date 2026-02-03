package testutil

import (
	"os"
	"runtime"
	"testing"
)

// TempDir returns a temporary directory with a short path on macOS.
func TempDir(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "darwin" {
		dir, err := os.MkdirTemp("/tmp", "pxcli-test-")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		t.Cleanup(func() {
			_ = os.RemoveAll(dir)
		})
		return dir
	}
	return t.TempDir()
}
