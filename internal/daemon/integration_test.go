package daemon

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pxcli/internal/config"
	"pxcli/internal/testutil"
)

func TestHeadlessProtocolIntegration(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")
	cfg := config.New(
		config.WithSocketPath(socketPath),
		config.WithPIDPath(pidPath),
		config.WithCanvasSize(4, 4),
	)

	done := startHeadlessRuntime(t, cfg)
	stopped := false
	t.Cleanup(func() {
		if stopped {
			return
		}
		if _, err := os.Stat(socketPath); err == nil {
			_, _ = sendRequest(socketPath, "stop\n")
		}
		assertRuntimeDone(t, done)
	})

	response := mustSendRequest(t, socketPath, "set_pixel -1 0 red\n")
	if !strings.HasPrefix(response, "err out_of_bounds ") {
		t.Fatalf("expected out_of_bounds error, got %q", response)
	}

	expectResponse(t, mustSendRequest(t, socketPath, "set_pixel 0 0 #ff0000\n"), "ok\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 0 0\n"), "ok #ff0000ff\n")

	expectResponse(t, mustSendRequest(t, socketPath, "fill_rect 1 1 2 2 blue\n"), "ok\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 1 1\n"), "ok #0000ffff\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 2 2\n"), "ok #0000ffff\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 3 3\n"), "ok #00000000\n")

	expectResponse(t, mustSendRequest(t, socketPath, "line 0 3 3 3 red\n"), "ok\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 0 3\n"), "ok #ff0000ff\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 3 3\n"), "ok #ff0000ff\n")

	expectResponse(t, mustSendRequest(t, socketPath, "clear\n"), "ok\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 0 0\n"), "ok #00000000\n")

	expectResponse(t, mustSendRequest(t, socketPath, "undo\n"), "ok\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 0 0\n"), "ok #ff0000ff\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 1 1\n"), "ok #0000ffff\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 0 3\n"), "ok #ff0000ff\n")

	exportPath := filepath.Join(dir, "out.png")
	expectResponse(t, mustSendRequest(t, socketPath, "export "+exportPath+"\n"), "ok\n")
	img := decodePNG(t, exportPath)
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Fatalf("expected 4x4 export, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
	assertPixel(t, img, 0, 0, 255, 0, 0, 255)
	assertPixel(t, img, 1, 1, 0, 0, 255, 255)
	assertPixel(t, img, 0, 3, 255, 0, 0, 255)

	expectResponse(t, mustSendRequest(t, socketPath, "redo\n"), "ok\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 0 0\n"), "ok #00000000\n")
	expectResponse(t, mustSendRequest(t, socketPath, "get_pixel 1 1\n"), "ok #00000000\n")

	expectResponse(t, mustSendRequest(t, socketPath, "stop\n"), "ok\n")
	assertRuntimeDone(t, done)
	assertPathMissing(t, pidPath)
	assertPathMissing(t, socketPath)
	stopped = true
}

func expectResponse(t *testing.T, response, expected string) {
	t.Helper()
	if response != expected {
		t.Fatalf("expected response %q, got %q", expected, response)
	}
}

func decodePNG(t *testing.T, path string) image.Image {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("unexpected png decode error: %v", err)
	}
	return img
}

func assertPixel(t *testing.T, img image.Image, x, y int, r, g, b, a uint8) {
	t.Helper()
	cr, cg, cb, ca := img.At(x, y).RGBA()
	if uint8(cr>>8) != r || uint8(cg>>8) != g || uint8(cb>>8) != b || uint8(ca>>8) != a {
		t.Fatalf("expected pixel (%d,%d) rgba(%d,%d,%d,%d), got rgba(%d,%d,%d,%d)",
			x, y, r, g, b, a, cr>>8, cg>>8, cb>>8, ca>>8)
	}
}
