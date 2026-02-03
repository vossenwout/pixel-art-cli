package canvas

import (
	"image/color"
	"testing"
)

func TestCanvasRenderSnapshotReflectsChanges(t *testing.T) {
	grid, err := New(2, 2)
	if err != nil {
		t.Fatalf("new canvas: %v", err)
	}
	colorValue := color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0x44}
	if err := grid.SetPixel(1, 0, colorValue); err != nil {
		t.Fatalf("set pixel: %v", err)
	}

	snapshot := grid.RenderSnapshot()
	if snapshot.Width != 2 || snapshot.Height != 2 {
		t.Fatalf("expected 2x2 snapshot, got %dx%d", snapshot.Width, snapshot.Height)
	}
	if len(snapshot.Pixels) != 2*2*4 {
		t.Fatalf("expected pixel buffer length %d, got %d", 2*2*4, len(snapshot.Pixels))
	}
	offset := (0*2 + 1) * 4
	got := snapshot.Pixels[offset : offset+4]
	want := []byte{colorValue.R, colorValue.G, colorValue.B, colorValue.A}
	for i, value := range want {
		if got[i] != value {
			t.Fatalf("expected pixel bytes %v, got %v", want, got)
		}
	}
}

func TestCanvasRenderSnapshotCopySemantics(t *testing.T) {
	grid, err := New(1, 1)
	if err != nil {
		t.Fatalf("new canvas: %v", err)
	}
	original := color.RGBA{R: 0xAA, G: 0xBB, B: 0xCC, A: 0xDD}
	if err := grid.SetPixel(0, 0, original); err != nil {
		t.Fatalf("set pixel: %v", err)
	}

	snapshot := grid.RenderSnapshot()
	snapshot.Pixels[0] = 0
	snapshot.Pixels[1] = 0
	snapshot.Pixels[2] = 0
	snapshot.Pixels[3] = 0

	got, err := grid.GetPixel(0, 0)
	if err != nil {
		t.Fatalf("get pixel: %v", err)
	}
	if got != original {
		t.Fatalf("expected canvas pixel %v, got %v", original, got)
	}
}

func TestCanvasDirtyFlagResetsAfterRenderSnapshot(t *testing.T) {
	grid, err := New(1, 1)
	if err != nil {
		t.Fatalf("new canvas: %v", err)
	}
	if err := grid.SetPixel(0, 0, color.RGBA{A: 0xFF}); err != nil {
		t.Fatalf("set pixel: %v", err)
	}
	if !grid.Dirty() {
		t.Fatalf("expected dirty after mutation")
	}

	_ = grid.RenderSnapshot()
	if grid.Dirty() {
		t.Fatalf("expected dirty to be false after render snapshot")
	}
}
