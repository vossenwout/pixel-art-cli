package protocol

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseLine(t *testing.T) {
	request, err := ParseLine("set_pixel 10 10 #ff0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request.Command != "set_pixel" {
		t.Fatalf("expected command set_pixel, got %q", request.Command)
	}
	wantArgs := []string{"10", "10", "#ff0000"}
	if !reflect.DeepEqual(request.Args, wantArgs) {
		t.Fatalf("expected args %v, got %v", wantArgs, request.Args)
	}
}

func TestParseLineWhitespace(t *testing.T) {
	request, err := ParseLine("  get_pixel   1\t2  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request.Command != "get_pixel" {
		t.Fatalf("expected command get_pixel, got %q", request.Command)
	}
	wantArgs := []string{"1", "2"}
	if !reflect.DeepEqual(request.Args, wantArgs) {
		t.Fatalf("expected args %v, got %v", wantArgs, request.Args)
	}
}

func TestParseLineEmpty(t *testing.T) {
	_, err := ParseLine("\t  \n")
	if err == nil {
		t.Fatalf("expected error")
	}
	var perr Error
	if !errors.As(err, &perr) {
		t.Fatalf("expected protocol Error, got %T", err)
	}
	if perr.Code != "invalid_command" {
		t.Fatalf("expected code invalid_command, got %q", perr.Code)
	}
}

func TestFormatOK(t *testing.T) {
	if got := FormatOK(""); got != "ok" {
		t.Fatalf("expected ok, got %q", got)
	}
	if got := FormatOK("#ff0000ff"); got != "ok #ff0000ff" {
		t.Fatalf("expected ok #ff0000ff, got %q", got)
	}
}

func TestFormatError(t *testing.T) {
	got := FormatError("invalid_command", "bad request")
	if got != "err invalid_command bad request" {
		t.Fatalf("expected err invalid_command bad request, got %q", got)
	}
}
