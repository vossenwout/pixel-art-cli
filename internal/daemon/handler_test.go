package daemon

import (
	"image/color"
	"strings"
	"testing"

	"pxcli/internal/canvas"
	pxcolor "pxcli/internal/color"
	"pxcli/internal/history"
	"pxcli/internal/protocol"
)

func TestHandlerGetPixel(t *testing.T) {
	target, err := canvas.New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := target.SetPixel(1, 2, color.RGBA{R: 1, G: 2, B: 3, A: 4}); err != nil {
		t.Fatalf("unexpected error setting pixel: %v", err)
	}
	handler := NewHandler(history.New(target), nil)

	response := handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "2"}})
	want := "ok " + pxcolor.Format(color.RGBA{R: 1, G: 2, B: 3, A: 4})
	if response != want {
		t.Fatalf("expected %q, got %q", want, response)
	}
}

func TestHandlerSetPixelArgCountError(t *testing.T) {
	target, err := canvas.New(2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	handler := NewHandler(history.New(target), nil)

	response := handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"1", "2"}})
	if !strings.HasPrefix(response, "err invalid_args ") {
		t.Fatalf("expected invalid_args error, got %q", response)
	}
}

func TestHandlerUnknownCommand(t *testing.T) {
	target, err := canvas.New(2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	handler := NewHandler(history.New(target), nil)

	response := handler.Handle(protocol.Request{Command: "nope"})
	if !strings.HasPrefix(response, "err invalid_command ") {
		t.Fatalf("expected invalid_command error, got %q", response)
	}
}

func TestHandlerMutatingCommandReturnsOK(t *testing.T) {
	target, err := canvas.New(2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	handler := NewHandler(history.New(target), nil)

	response := handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "1", "red"}})
	if response != "ok" {
		t.Fatalf("expected ok, got %q", response)
	}
	value, err := target.GetPixel(0, 1)
	if err != nil {
		t.Fatalf("unexpected error reading pixel: %v", err)
	}
	if value != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected red pixel, got %+v", value)
	}
}

func TestHandlerClearDefaultTransparent(t *testing.T) {
	target, err := canvas.New(2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := target.SetPixel(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255}); err != nil {
		t.Fatalf("unexpected error setting pixel: %v", err)
	}
	handler := NewHandler(history.New(target), nil)

	response := handler.Handle(protocol.Request{Command: "clear"})
	if response != "ok" {
		t.Fatalf("expected ok, got %q", response)
	}
	value, err := target.GetPixel(0, 0)
	if err != nil {
		t.Fatalf("unexpected error reading pixel: %v", err)
	}
	if value != (color.RGBA{R: 0, G: 0, B: 0, A: 0}) {
		t.Fatalf("expected transparent pixel, got %+v", value)
	}
}

func TestHandlerStopTriggersCallback(t *testing.T) {
	target, err := canvas.New(2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	called := false
	handler := NewHandler(history.New(target), func() {
		called = true
	})

	response := handler.Handle(protocol.Request{Command: "stop"})
	if response != "ok" {
		t.Fatalf("expected ok, got %q", response)
	}
	if !called {
		t.Fatalf("expected stop callback to be invoked")
	}
}
