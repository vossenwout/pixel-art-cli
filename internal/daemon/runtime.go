package daemon

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Stopper provides a reusable stop signal for coordinating shutdown.
type Stopper struct {
	once sync.Once
	ch   chan struct{}
}

// NewStopper creates a new Stopper.
func NewStopper() *Stopper {
	return &Stopper{ch: make(chan struct{})}
}

// Stop closes the stop channel once.
func (s *Stopper) Stop() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		close(s.ch)
	})
}

// Done returns a channel that is closed when Stop is called.
func (s *Stopper) Done() <-chan struct{} {
	if s == nil {
		return nil
	}
	return s.ch
}

// RuntimeOptions configures the daemon runtime lifecycle.
type RuntimeOptions struct {
	PIDPath    string
	SocketPath string
	StopCh     <-chan struct{}
	SignalCh   <-chan os.Signal
}

// Runtime coordinates server lifecycle, stop requests, and cleanup.
type Runtime struct {
	server     *Server
	pidPath    string
	socketPath string
	stopCh     <-chan struct{}
	signalCh   <-chan os.Signal
	stopOnce   sync.Once
}

// NewRuntime creates a runtime for the provided server.
func NewRuntime(server *Server, opts RuntimeOptions) (*Runtime, error) {
	if server == nil {
		return nil, errors.New("server must not be nil")
	}
	return &Runtime{
		server:     server,
		pidPath:    opts.PIDPath,
		socketPath: opts.SocketPath,
		stopCh:     opts.StopCh,
		signalCh:   opts.SignalCh,
	}, nil
}

// Run blocks until the server stops, a stop request arrives, or a signal is received.
func (r *Runtime) Run() error {
	if r.server == nil {
		return errors.New("server must not be nil")
	}
	stopCh := r.stopCh
	if stopCh == nil {
		stopCh = make(chan struct{})
	}

	signalCh := r.signalCh
	var stopSignals func()
	if signalCh == nil {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		signalCh = ch
		stopSignals = func() {
			signal.Stop(ch)
		}
	}
	if stopSignals != nil {
		defer stopSignals()
	}

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- r.server.Serve()
	}()

	var serveErr error
	select {
	case serveErr = <-serverDone:
	case <-stopCh:
		r.Stop()
		serveErr = <-serverDone
	case <-signalCh:
		r.Stop()
		serveErr = <-serverDone
	}

	cleanupErr := CleanupFiles(r.pidPath, r.socketPath)
	if serveErr != nil {
		return serveErr
	}
	if cleanupErr != nil {
		return cleanupErr
	}
	return nil
}

// Stop closes the server listener once.
func (r *Runtime) Stop() {
	if r == nil || r.server == nil {
		return
	}
	r.stopOnce.Do(func() {
		_ = r.server.Close()
	})
}
