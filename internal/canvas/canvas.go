package canvas

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"
)

// Error represents a canvas error with a code and message.
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

// Canvas stores pixel data for a fixed-width, fixed-height image.
type Canvas struct {
	mu     sync.RWMutex
	width  int
	height int
	pixels []color.RGBA
}

// Snapshot captures a copy of the canvas pixels.
type Snapshot struct {
	width  int
	height int
	pixels []color.RGBA
}

// New creates a canvas with the provided dimensions.
func New(width, height int) (*Canvas, error) {
	if width <= 0 || height <= 0 {
		return nil, Error{Code: "invalid_args", Message: "canvas dimensions must be positive"}
	}
	pixels := make([]color.RGBA, width*height)
	return &Canvas{width: width, height: height, pixels: pixels}, nil
}

// Width returns the canvas width in pixels.
func (c *Canvas) Width() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.width
}

// Height returns the canvas height in pixels.
func (c *Canvas) Height() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.height
}

// SetPixel sets a pixel to the provided color.
func (c *Canvas) SetPixel(x, y int, value color.RGBA) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	idx, err := c.index(x, y)
	if err != nil {
		return err
	}
	c.pixels[idx] = value
	return nil
}

// GetPixel returns the color at the provided coordinates.
func (c *Canvas) GetPixel(x, y int) (color.RGBA, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	idx, err := c.index(x, y)
	if err != nil {
		return color.RGBA{}, err
	}
	return c.pixels[idx], nil
}

// Clear fills the entire canvas with the provided color.
func (c *Canvas) Clear(value color.RGBA) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.pixels {
		c.pixels[i] = value
	}
}

// Snapshot returns a copy of the current canvas state.
func (c *Canvas) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pixels := make([]color.RGBA, len(c.pixels))
	copy(pixels, c.pixels)
	return Snapshot{width: c.width, height: c.height, pixels: pixels}
}

// Restore replaces the current canvas state with the snapshot.
func (c *Canvas) Restore(snapshot Snapshot) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if snapshot.width != c.width || snapshot.height != c.height {
		return Error{Code: "invalid_args", Message: "snapshot dimensions do not match canvas"}
	}
	if len(snapshot.pixels) != len(c.pixels) {
		return Error{Code: "invalid_args", Message: "snapshot size does not match canvas"}
	}
	copy(c.pixels, snapshot.pixels)
	return nil
}

// FillRect fills a rectangle with the provided color.
func (c *Canvas) FillRect(x, y, w, h int, value color.RGBA) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if w <= 0 || h <= 0 {
		return Error{Code: "invalid_args", Message: "rect width and height must be positive"}
	}
	if x < 0 || y < 0 || x+w > c.width || y+h > c.height {
		return Error{
			Code:    "out_of_bounds",
			Message: fmt.Sprintf("rect (%d,%d) size %dx%d outside canvas", x, y, w, h),
		}
	}

	for row := y; row < y+h; row++ {
		start := row*c.width + x
		for i := 0; i < w; i++ {
			c.pixels[start+i] = value
		}
	}
	return nil
}

// Line draws a line between two points, inclusive of endpoints.
func (c *Canvas) Line(x1, y1, x2, y2 int, value color.RGBA) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.index(x1, y1); err != nil {
		return err
	}
	if _, err := c.index(x2, y2); err != nil {
		return err
	}

	dx := absInt(x2 - x1)
	dy := absInt(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	errVal := dx - dy

	for {
		c.pixels[y1*c.width+x1] = value
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * errVal
		if e2 > -dy {
			errVal -= dy
			x1 += sx
		}
		if e2 < dx {
			errVal += dx
			y1 += sy
		}
	}
	return nil
}

// ExportPNG writes the canvas to a PNG file at the provided path.
func (c *Canvas) ExportPNG(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return Error{Code: "io", Message: err.Error()}
	}

	snapshot := c.Snapshot()
	img := image.NewRGBA(image.Rect(0, 0, snapshot.width, snapshot.height))
	for y := 0; y < snapshot.height; y++ {
		row := y * snapshot.width
		for x := 0; x < snapshot.width; x++ {
			img.SetRGBA(x, y, snapshot.pixels[row+x])
		}
	}

	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		return Error{Code: "io", Message: err.Error()}
	}
	if err := file.Close(); err != nil {
		return Error{Code: "io", Message: err.Error()}
	}
	return nil
}

func (c *Canvas) index(x, y int) (int, error) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return 0, Error{Code: "out_of_bounds", Message: fmt.Sprintf("pixel (%d,%d) outside canvas", x, y)}
	}
	return y*c.width + x, nil
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
