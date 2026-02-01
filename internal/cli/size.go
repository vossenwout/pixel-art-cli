package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func parseCanvasSize(raw string) (int, int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, 0, fmt.Errorf("invalid size %q: expected WxH", raw)
	}
	parts := strings.Split(strings.ToLower(trimmed), "x")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, fmt.Errorf("invalid size %q: expected WxH", raw)
	}
	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid size %q: expected WxH", raw)
	}
	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid size %q: expected WxH", raw)
	}
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid size %q: width and height must be positive", raw)
	}
	return width, height, nil
}
