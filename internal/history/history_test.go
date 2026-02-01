package history

import (
	"image/color"
	"testing"

	"pxcli/internal/canvas"
)

func TestHistoryUndoRedoRestoresState(t *testing.T) {
	c, err := canvas.New(2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manager := New(c)
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	if err := manager.Apply(func(target *canvas.Canvas) error {
		return target.SetPixel(1, 1, red)
	}); err != nil {
		t.Fatalf("unexpected apply error: %v", err)
	}

	got, err := c.GetPixel(1, 1)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if got != red {
		t.Fatalf("expected red before undo, got %v", got)
	}

	if err := manager.Undo(); err != nil {
		t.Fatalf("unexpected undo error: %v", err)
	}

	got, err = c.GetPixel(1, 1)
	if err != nil {
		t.Fatalf("unexpected get error after undo: %v", err)
	}
	if got != (color.RGBA{}) {
		t.Fatalf("expected zero color after undo, got %v", got)
	}

	if err := manager.Redo(); err != nil {
		t.Fatalf("unexpected redo error: %v", err)
	}

	got, err = c.GetPixel(1, 1)
	if err != nil {
		t.Fatalf("unexpected get error after redo: %v", err)
	}
	if got != red {
		t.Fatalf("expected red after redo, got %v", got)
	}
}

func TestHistoryNoHistoryErrors(t *testing.T) {
	c, err := canvas.New(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manager := New(c)

	if err := manager.Undo(); err == nil {
		t.Fatalf("expected undo error")
	} else if historyErr, ok := err.(Error); !ok || historyErr.Code != "no_history" {
		t.Fatalf("expected no_history error, got %v", err)
	}

	if err := manager.Redo(); err == nil {
		t.Fatalf("expected redo error")
	} else if historyErr, ok := err.(Error); !ok || historyErr.Code != "no_history" {
		t.Fatalf("expected no_history error, got %v", err)
	}
}

func TestHistoryRedoClearedOnNewMutation(t *testing.T) {
	c, err := canvas.New(2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manager := New(c)
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	if err := manager.Apply(func(target *canvas.Canvas) error {
		return target.SetPixel(0, 0, red)
	}); err != nil {
		t.Fatalf("unexpected apply error: %v", err)
	}

	if err := manager.Undo(); err != nil {
		t.Fatalf("unexpected undo error: %v", err)
	}

	if err := manager.Apply(func(target *canvas.Canvas) error {
		return target.SetPixel(1, 0, green)
	}); err != nil {
		t.Fatalf("unexpected apply error: %v", err)
	}

	if err := manager.Redo(); err == nil {
		t.Fatalf("expected redo error after new mutation")
	} else if historyErr, ok := err.(Error); !ok || historyErr.Code != "no_history" {
		t.Fatalf("expected no_history error, got %v", err)
	}

	got, err := c.GetPixel(1, 0)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if got != green {
		t.Fatalf("expected green at (1,0), got %v", got)
	}
}
