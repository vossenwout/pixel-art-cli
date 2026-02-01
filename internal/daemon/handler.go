package daemon

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"

	"pxcli/internal/canvas"
	pxcolor "pxcli/internal/color"
	"pxcli/internal/history"
	"pxcli/internal/protocol"
)

// Handler maps protocol requests to canvas operations.
type Handler struct {
	history *history.Manager
	onStop  func()
}

// NewHandler creates a command handler for the provided history manager.
func NewHandler(history *history.Manager, onStop func()) *Handler {
	return &Handler{history: history, onStop: onStop}
}

// Handle executes a command and returns a single-line protocol response.
func (h *Handler) Handle(request protocol.Request) string {
	switch request.Command {
	case "set_pixel":
		return h.handleSetPixel(request.Args)
	case "get_pixel":
		return h.handleGetPixel(request.Args)
	case "fill_rect":
		return h.handleFillRect(request.Args)
	case "line":
		return h.handleLine(request.Args)
	case "clear":
		return h.handleClear(request.Args)
	case "export":
		return h.handleExport(request.Args)
	case "undo":
		return h.handleUndo(request.Args)
	case "redo":
		return h.handleRedo(request.Args)
	case "stop":
		return h.handleStop(request.Args)
	default:
		return protocol.FormatError("invalid_command", fmt.Sprintf("unknown command %q", request.Command))
	}
}

func (h *Handler) handleSetPixel(args []string) string {
	if len(args) != 3 {
		return invalidArgCount(3, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return formatError(err)
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return formatError(err)
	}
	value, err := pxcolor.Parse(args[2])
	if err != nil {
		return formatError(err)
	}
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return c.SetPixel(x, y, value)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleGetPixel(args []string) string {
	if len(args) != 2 {
		return invalidArgCount(2, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return formatError(err)
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return formatError(err)
	}
	value, err := h.history.Canvas().GetPixel(x, y)
	if err != nil {
		return formatError(err)
	}
	return protocol.FormatOK(pxcolor.Format(value))
}

func (h *Handler) handleFillRect(args []string) string {
	if len(args) != 5 {
		return invalidArgCount(5, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return formatError(err)
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return formatError(err)
	}
	w, err := parseIntArg(args[2], "w")
	if err != nil {
		return formatError(err)
	}
	hgt, err := parseIntArg(args[3], "h")
	if err != nil {
		return formatError(err)
	}
	value, err := pxcolor.Parse(args[4])
	if err != nil {
		return formatError(err)
	}
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return c.FillRect(x, y, w, hgt, value)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleLine(args []string) string {
	if len(args) != 5 {
		return invalidArgCount(5, len(args))
	}
	x1, err := parseIntArg(args[0], "x1")
	if err != nil {
		return formatError(err)
	}
	y1, err := parseIntArg(args[1], "y1")
	if err != nil {
		return formatError(err)
	}
	x2, err := parseIntArg(args[2], "x2")
	if err != nil {
		return formatError(err)
	}
	y2, err := parseIntArg(args[3], "y2")
	if err != nil {
		return formatError(err)
	}
	value, err := pxcolor.Parse(args[4])
	if err != nil {
		return formatError(err)
	}
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return c.Line(x1, y1, x2, y2, value)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleClear(args []string) string {
	if len(args) > 1 {
		return invalidArgCount(1, len(args))
	}
	value := canvasTransparent()
	if len(args) == 1 {
		parsed, err := pxcolor.Parse(args[0])
		if err != nil {
			return formatError(err)
		}
		value = parsed
	}
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		c.Clear(value)
		return nil
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleExport(args []string) string {
	if len(args) != 1 {
		return invalidArgCount(1, len(args))
	}
	if err := h.history.Canvas().ExportPNG(args[0]); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleUndo(args []string) string {
	if len(args) != 0 {
		return invalidArgCount(0, len(args))
	}
	if err := h.history.Undo(); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleRedo(args []string) string {
	if len(args) != 0 {
		return invalidArgCount(0, len(args))
	}
	if err := h.history.Redo(); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleStop(args []string) string {
	if len(args) != 0 {
		return invalidArgCount(0, len(args))
	}
	if h.onStop != nil {
		h.onStop()
	}
	return protocol.FormatOK("")
}

func parseIntArg(value, name string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, handlerError{Code: "invalid_args", Message: fmt.Sprintf("%s must be an integer", name)}
	}
	return parsed, nil
}

func invalidArgCount(expected, got int) string {
	return protocol.FormatError("invalid_args", fmt.Sprintf("expected %d args, got %d", expected, got))
}

func canvasTransparent() color.RGBA {
	return color.RGBA{R: 0, G: 0, B: 0, A: 0}
}

type handlerError struct {
	Code    string
	Message string
}

func (e handlerError) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

func formatError(err error) string {
	if err == nil {
		return protocol.FormatError("error", "unknown error")
	}
	var herr handlerError
	if errors.As(err, &herr) {
		return protocol.FormatError(herr.Code, herr.Message)
	}
	var cerr canvas.Error
	if errors.As(err, &cerr) {
		return protocol.FormatError(cerr.Code, cerr.Message)
	}
	var colErr pxcolor.Error
	if errors.As(err, &colErr) {
		return protocol.FormatError(colErr.Code, colErr.Message)
	}
	var histErr history.Error
	if errors.As(err, &histErr) {
		return protocol.FormatError(histErr.Code, histErr.Message)
	}
	return protocol.FormatError("error", err.Error())
}
