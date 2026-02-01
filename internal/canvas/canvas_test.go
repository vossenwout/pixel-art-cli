package canvas

import (
	"image/color"
	"testing"
)

func TestCanvasSetGetPixel(t *testing.T) {
	c, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if err := c.SetPixel(1, 2, red); err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	got, err := c.GetPixel(1, 2)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if got != red {
		t.Fatalf("expected %v, got %v", red, got)
	}
}

func TestCanvasBounds(t *testing.T) {
	c, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	badCoords := [][2]int{
		{-1, 0},
		{0, -1},
		{4, 0},
		{0, 4},
	}

	for _, coords := range badCoords {
		x, y := coords[0], coords[1]
		if err := c.SetPixel(x, y, color.RGBA{}); err == nil {
			t.Fatalf("expected set error for (%d,%d)", x, y)
		} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
			t.Fatalf("expected out_of_bounds for (%d,%d), got %v", x, y, err)
		}

		if _, err := c.GetPixel(x, y); err == nil {
			t.Fatalf("expected get error for (%d,%d)", x, y)
		} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
			t.Fatalf("expected out_of_bounds for (%d,%d), got %v", x, y, err)
		}
	}
}

func TestCanvasClear(t *testing.T) {
	c, err := New(3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	c.Clear(blue)

	for y := 0; y < c.Height(); y++ {
		for x := 0; x < c.Width(); x++ {
			got, err := c.GetPixel(x, y)
			if err != nil {
				t.Fatalf("unexpected get error at (%d,%d): %v", x, y, err)
			}
			if got != blue {
				t.Fatalf("expected %v at (%d,%d), got %v", blue, x, y, got)
			}
		}
	}
}
