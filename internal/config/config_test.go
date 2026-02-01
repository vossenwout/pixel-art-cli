package config

import "testing"

func TestDefaultConfigValues(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.SocketPath != DefaultSocketPath {
		t.Fatalf("expected default socket path %q, got %q", DefaultSocketPath, cfg.SocketPath)
	}
	if cfg.PIDPath != DefaultPIDPath {
		t.Fatalf("expected default PID path %q, got %q", DefaultPIDPath, cfg.PIDPath)
	}
	if cfg.CanvasWidth != DefaultCanvasWidth || cfg.CanvasHeight != DefaultCanvasHeight {
		t.Fatalf("expected default canvas size %dx%d, got %dx%d", DefaultCanvasWidth, DefaultCanvasHeight, cfg.CanvasWidth, cfg.CanvasHeight)
	}
	if cfg.Scale != DefaultScale {
		t.Fatalf("expected default scale %d, got %d", DefaultScale, cfg.Scale)
	}
	if cfg.Headless != DefaultHeadless {
		t.Fatalf("expected default headless %v, got %v", DefaultHeadless, cfg.Headless)
	}
}

func TestConfigOverrides(t *testing.T) {
	cfg := New(
		WithSocketPath("/tmp/test.sock"),
		WithPIDPath("/tmp/test.pid"),
		WithCanvasSize(8, 9),
		WithScale(3),
		WithHeadless(false),
	)

	if cfg.SocketPath != "/tmp/test.sock" {
		t.Fatalf("expected socket override, got %q", cfg.SocketPath)
	}
	if cfg.PIDPath != "/tmp/test.pid" {
		t.Fatalf("expected PID override, got %q", cfg.PIDPath)
	}
	if cfg.CanvasWidth != 8 || cfg.CanvasHeight != 9 {
		t.Fatalf("expected canvas override 8x9, got %dx%d", cfg.CanvasWidth, cfg.CanvasHeight)
	}
	if cfg.Scale != 3 {
		t.Fatalf("expected scale override 3, got %d", cfg.Scale)
	}
	if cfg.Headless != false {
		t.Fatalf("expected headless override false, got %v", cfg.Headless)
	}
}
