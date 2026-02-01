package canvas

import (
	"fmt"
	"image/color"
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
	return c.width
}

// Height returns the canvas height in pixels.
func (c *Canvas) Height() int {
	return c.height
}

// SetPixel sets a pixel to the provided color.
func (c *Canvas) SetPixel(x, y int, value color.RGBA) error {
	idx, err := c.index(x, y)
	if err != nil {
		return err
	}
	c.pixels[idx] = value
	return nil
}

// GetPixel returns the color at the provided coordinates.
func (c *Canvas) GetPixel(x, y int) (color.RGBA, error) {
	idx, err := c.index(x, y)
	if err != nil {
		return color.RGBA{}, err
	}
	return c.pixels[idx], nil
}

// Clear fills the entire canvas with the provided color.
func (c *Canvas) Clear(value color.RGBA) {
	for i := range c.pixels {
		c.pixels[i] = value
	}
}

// FillRect fills a rectangle with the provided color.
func (c *Canvas) FillRect(x, y, w, h int, value color.RGBA) error {
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

func (c *Canvas) index(x, y int) (int, error) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return 0, Error{Code: "out_of_bounds", Message: fmt.Sprintf("pixel (%d,%d) outside canvas", x, y)}
	}
	return y*c.width + x, nil
}
