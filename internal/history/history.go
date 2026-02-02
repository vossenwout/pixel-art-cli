package history

import (
	"sync"

	"pxcli/internal/canvas"
)

// Error represents an undo/redo history error with a code and message.
type Error struct {
	Code    string
	Message string
}

func (e Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

// Manager tracks undo/redo history for a canvas.
type Manager struct {
	mu     sync.Mutex
	canvas *canvas.Canvas
	undo   []canvas.Snapshot
	redo   []canvas.Snapshot
}

// New creates a new history manager for the provided canvas.
func New(target *canvas.Canvas) *Manager {
	return &Manager{canvas: target}
}

// Canvas returns the managed canvas.
func (m *Manager) Canvas() *canvas.Canvas {
	return m.canvas
}

// Apply runs a mutating operation and records undo history on success.
func (m *Manager) Apply(mutate func(*canvas.Canvas) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	snapshot := m.canvas.Snapshot()
	if err := mutate(m.canvas); err != nil {
		return err
	}
	m.undo = append(m.undo, snapshot)
	m.redo = nil
	return nil
}

// Undo restores the previous snapshot, if available.
func (m *Manager) Undo() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.undo) == 0 {
		return Error{Code: "no_history", Message: "nothing to undo"}
	}
	current := m.canvas.Snapshot()
	previous := m.undo[len(m.undo)-1]
	if err := m.canvas.Restore(previous); err != nil {
		return err
	}
	m.undo = m.undo[:len(m.undo)-1]
	m.redo = append(m.redo, current)
	return nil
}

// Redo reapplies a previously undone snapshot, if available.
func (m *Manager) Redo() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.redo) == 0 {
		return Error{Code: "no_history", Message: "nothing to redo"}
	}
	current := m.canvas.Snapshot()
	next := m.redo[len(m.redo)-1]
	if err := m.canvas.Restore(next); err != nil {
		return err
	}
	m.redo = m.redo[:len(m.redo)-1]
	m.undo = append(m.undo, current)
	return nil
}
