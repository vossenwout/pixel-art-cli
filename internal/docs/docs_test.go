package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeCoversCoreWorkflow(t *testing.T) {
	readmePath := filepath.Join("..", "..", "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README: %v", err)
	}
	content := strings.ToLower(string(data))

	required := []string{
		"pxcli start",
		"pxcli stop",
		"set_pixel",
		"get_pixel",
		"export",
		"undo",
		"redo",
		"headless",
		"-tags=ebiten",
		"headless container",
		"--socket",
		"stale pid/socket",
		"absolute path",
		"export <filename.png>",
	}

	for _, term := range required {
		if !strings.Contains(content, term) {
			t.Fatalf("README missing %q", term)
		}
	}
}
