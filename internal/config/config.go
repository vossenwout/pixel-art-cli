package config

const (
	DefaultSocketPath   = "/tmp/pxcli.sock"
	DefaultPIDPath      = "/tmp/pxcli.pid"
	DefaultCanvasWidth  = 32
	DefaultCanvasHeight = 32
	DefaultScale        = 10
	DefaultHeadless     = false
)

// Config holds shared defaults and overrides for CLI and daemon behavior.
type Config struct {
	SocketPath   string
	PIDPath      string
	CanvasWidth  int
	CanvasHeight int
	Scale        int
	Headless     bool
}

// DefaultConfig returns the default configuration values.
func DefaultConfig() Config {
	return Config{
		SocketPath:   DefaultSocketPath,
		PIDPath:      DefaultPIDPath,
		CanvasWidth:  DefaultCanvasWidth,
		CanvasHeight: DefaultCanvasHeight,
		Scale:        DefaultScale,
		Headless:     DefaultHeadless,
	}
}

// Option customizes configuration values.
type Option func(*Config)

// New creates a config with defaults, then applies any overrides.
func New(opts ...Option) Config {
	cfg := DefaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// WithSocketPath overrides the socket path.
func WithSocketPath(path string) Option {
	return func(cfg *Config) {
		cfg.SocketPath = path
	}
}

// WithPIDPath overrides the PID file path.
func WithPIDPath(path string) Option {
	return func(cfg *Config) {
		cfg.PIDPath = path
	}
}

// WithCanvasSize overrides the default canvas size.
func WithCanvasSize(width, height int) Option {
	return func(cfg *Config) {
		cfg.CanvasWidth = width
		cfg.CanvasHeight = height
	}
}

// WithScale overrides the default scale.
func WithScale(scale int) Option {
	return func(cfg *Config) {
		cfg.Scale = scale
	}
}

// WithHeadless overrides the default headless setting.
func WithHeadless(headless bool) Option {
	return func(cfg *Config) {
		cfg.Headless = headless
	}
}
