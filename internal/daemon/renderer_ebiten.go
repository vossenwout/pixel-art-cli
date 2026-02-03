//go:build ebiten

package daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

var errRendererClosed = errors.New("renderer closed")

// RendererAvailable reports whether GUI rendering is supported in this build.
func RendererAvailable() bool {
	return true
}

// EbitenRenderer renders the canvas with Ebiten.
type EbitenRenderer struct {
	source    RenderSource
	scale     int
	closeOnce sync.Once
	closeCh   chan struct{}
}

// NewEbitenRenderer creates a renderer for the provided render source.
func NewEbitenRenderer(source RenderSource, scale int) (*EbitenRenderer, error) {
	if source == nil {
		return nil, errors.New("render source must not be nil")
	}
	if scale <= 0 {
		return nil, fmt.Errorf("scale must be > 0, got %d", scale)
	}
	return &EbitenRenderer{
		source:  source,
		scale:   scale,
		closeCh: make(chan struct{}),
	}, nil
}

// Run starts the Ebiten render loop on the current goroutine.
func (r *EbitenRenderer) Run(ctx context.Context) error {
	if r == nil {
		return errors.New("renderer must not be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	width, height := r.source.Width(), r.source.Height()
	windowW, windowH := scaledWindowSize(width, height, r.scale)
	ebiten.SetWindowTitle("pxcli")
	ebiten.SetWindowSize(windowW, windowH)
	ebiten.SetWindowResizable(false)

	game := &renderGame{
		source:  r.source,
		scale:   r.scale,
		width:   width,
		height:  height,
		closeCh: r.closeCh,
		ctx:     ctx,
	}
	if err := ebiten.RunGame(game); err != nil {
		if errors.Is(err, errRendererClosed) {
			return nil
		}
		return err
	}
	return nil
}

// RequestClose signals the renderer to exit on the next update tick.
func (r *EbitenRenderer) RequestClose() {
	if r == nil {
		return
	}
	r.closeOnce.Do(func() {
		close(r.closeCh)
	})
}

type renderGame struct {
	source  RenderSource
	scale   int
	width   int
	height  int
	img     *ebiten.Image
	closeCh <-chan struct{}
	ctx     context.Context
}

func (g *renderGame) Update() error {
	select {
	case <-g.closeCh:
		return errRendererClosed
	default:
	}

	if g.ctx != nil {
		select {
		case <-g.ctx.Done():
			return errRendererClosed
		default:
		}
	}

	if g.img == nil || g.source.Dirty() {
		snapshot := g.source.RenderSnapshot()
		if g.img == nil {
			g.img = ebiten.NewImage(snapshot.Width, snapshot.Height)
			g.width = snapshot.Width
			g.height = snapshot.Height
		}
		g.img.ReplacePixels(snapshot.Pixels)
	}
	return nil
}

func (g *renderGame) Draw(screen *ebiten.Image) {
	if g.img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(g.scale), float64(g.scale))
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(g.img, op)
}

func (g *renderGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.width == 0 || g.height == 0 {
		return outsideWidth, outsideHeight
	}
	return scaledWindowSize(g.width, g.height, g.scale)
}
