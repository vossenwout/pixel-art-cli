package color

import (
	"fmt"
	"image/color"
	"strings"
)

// Error represents a color parsing error with a code and message.
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

// Parse converts a color string into RGBA.
// Supported formats: #rgb, #rrggbb, #rrggbbaa, and named colors red/blue/transparent.
func Parse(input string) (color.RGBA, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return color.RGBA{}, Error{Code: "invalid_color", Message: "color is required"}
	}

	lower := strings.ToLower(trimmed)
	switch lower {
	case "red":
		return color.RGBA{R: 255, G: 0, B: 0, A: 255}, nil
	case "blue":
		return color.RGBA{R: 0, G: 0, B: 255, A: 255}, nil
	case "transparent":
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}, nil
	}

	if !strings.HasPrefix(lower, "#") {
		return color.RGBA{}, Error{Code: "invalid_color", Message: "expected hex color or named color"}
	}

	hex := lower[1:]
	switch len(hex) {
	case 3:
		r, ok := hexNibble(hex[0])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		g, ok := hexNibble(hex[1])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		b, ok := hexNibble(hex[2])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		return color.RGBA{R: r * 17, G: g * 17, B: b * 17, A: 255}, nil
	case 6:
		r, ok := hexByte(hex[0:2])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		g, ok := hexByte(hex[2:4])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		b, ok := hexByte(hex[4:6])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		return color.RGBA{R: r, G: g, B: b, A: 255}, nil
	case 8:
		r, ok := hexByte(hex[0:2])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		g, ok := hexByte(hex[2:4])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		b, ok := hexByte(hex[4:6])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		a, ok := hexByte(hex[6:8])
		if !ok {
			return color.RGBA{}, invalidHexError(trimmed)
		}
		return color.RGBA{R: r, G: g, B: b, A: a}, nil
	default:
		return color.RGBA{}, Error{Code: "invalid_color", Message: fmt.Sprintf("invalid hex length %d", len(hex))}
	}
}

// Format returns the canonical #rrggbbaa lowercase format for a color.
func Format(c color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}

func hexNibble(b byte) (uint8, bool) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', true
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, true
	default:
		return 0, false
	}
}

func hexByte(s string) (uint8, bool) {
	if len(s) != 2 {
		return 0, false
	}
	hi, ok := hexNibble(s[0])
	if !ok {
		return 0, false
	}
	lo, ok := hexNibble(s[1])
	if !ok {
		return 0, false
	}
	return hi<<4 | lo, true
}

func invalidHexError(input string) Error {
	return Error{Code: "invalid_color", Message: fmt.Sprintf("invalid hex color %q", input)}
}
