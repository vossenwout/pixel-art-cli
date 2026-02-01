package canvas

import (
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCanvasExportPNG(t *testing.T) {
	c, err := New(2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	transparent := color.RGBA{R: 0, G: 0, B: 0, A: 0}

	if err := c.SetPixel(0, 0, red); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}
	if err := c.SetPixel(1, 0, green); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}
	if err := c.SetPixel(0, 1, blue); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}
	if err := c.SetPixel(1, 1, transparent); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	path := filepath.Join(t.TempDir(), "out.png")
	if err := c.ExportPNG(path); err != nil {
		t.Fatalf("unexpected export error: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Fatalf("expected 2x2 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	assertPixel := func(x, y int, want color.RGBA) {
		got := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
		if got != want {
			t.Fatalf("expected %v at (%d,%d), got %v", want, x, y, got)
		}
	}

	assertPixel(0, 0, red)
	assertPixel(1, 0, green)
	assertPixel(0, 1, blue)
	assertPixel(1, 1, transparent)
}

func TestCanvasExportPNGIOError(t *testing.T) {
	c, err := New(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(t.TempDir(), "missing", "out.png")
	if err := c.ExportPNG(path); err == nil {
		t.Fatalf("expected io error")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "io" {
		t.Fatalf("expected io error, got %v", err)
	}
}
