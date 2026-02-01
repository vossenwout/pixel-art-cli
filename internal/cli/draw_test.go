package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"pxcli/internal/client"
)

type stubClient struct {
	requests []string
	response client.Response
	err      error
}

func (s *stubClient) Send(request string) (client.Response, error) {
	s.requests = append(s.requests, request)
	return s.response, s.err
}

func TestDrawCommands_FormatRequests(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantRequest string
	}{
		{
			name:        "set_pixel",
			args:        []string{"set_pixel", "1", "2", "red"},
			wantRequest: "set_pixel 1 2 red",
		},
		{
			name:        "fill_rect",
			args:        []string{"fill_rect", "0", "1", "2", "3", "#ff00ff"},
			wantRequest: "fill_rect 0 1 2 3 #ff00ff",
		},
		{
			name:        "line",
			args:        []string{"line", "0", "0", "3", "0", "blue"},
			wantRequest: "line 0 0 3 0 blue",
		},
		{
			name:        "clear",
			args:        []string{"clear"},
			wantRequest: "clear",
		},
		{
			name:        "clear_with_color",
			args:        []string{"clear", "transparent"},
			wantRequest: "clear transparent",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubClient{
				response: client.Response{Raw: "ok"},
			}
			restore := drawNewClient
			drawNewClient = func(socketPath string) (requestSender, error) {
				return stub, nil
			}
			t.Cleanup(func() {
				drawNewClient = restore
			})

			buf := &bytes.Buffer{}
			cmd := NewRootCmd("dev")
			cmd.SetOut(buf)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(stub.requests) != 1 {
				t.Fatalf("expected 1 request, got %d", len(stub.requests))
			}
			if stub.requests[0] != tt.wantRequest {
				t.Fatalf("expected request %q, got %q", tt.wantRequest, stub.requests[0])
			}
			if strings.TrimSpace(buf.String()) != "ok" {
				t.Fatalf("expected ok output, got %q", buf.String())
			}
		})
	}
}

func TestFillRectCmd_InvalidSize(t *testing.T) {
	called := false
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		called = true
		return nil, fmt.Errorf("client should not be created")
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"fill_rect", "0", "0", "-1", "2", "red"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid size")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestLineCmd_InvalidInteger(t *testing.T) {
	called := false
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		called = true
		return nil, fmt.Errorf("client should not be created")
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"line", "0", "a", "1", "2", "red"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid integer")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}
