package color

import (
	"errors"
	"image/color"
	"testing"
)

func TestParseShortHex(t *testing.T) {
	got, err := Parse("#f00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParseInvalidColor(t *testing.T) {
	_, err := Parse("#12")
	if err == nil {
		t.Fatalf("expected error")
	}
	var cerr Error
	if !errors.As(err, &cerr) {
		t.Fatalf("expected color Error, got %T", err)
	}
	if cerr.Code != "invalid_color" {
		t.Fatalf("expected code invalid_color, got %q", cerr.Code)
	}

	_, err = Parse("not-a-color")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.As(err, &cerr) {
		t.Fatalf("expected color Error, got %T", err)
	}
	if cerr.Code != "invalid_color" {
		t.Fatalf("expected code invalid_color, got %q", cerr.Code)
	}
}

func TestFormatTransparent(t *testing.T) {
	got := Format(color.RGBA{R: 0, G: 0, B: 0, A: 0})
	if got != "#00000000" {
		t.Fatalf("expected #00000000, got %q", got)
	}
}
