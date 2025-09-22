package screen

import (
	"image"
	"image/color"
	"nonogram/bitmask"
)

var (
	_ image.Image = (*OneBitImage)(nil)
)

func NewOneBitImage(r image.Rectangle, threshold float64) *OneBitImage {
	width := r.Dx()
	height := r.Dy()
	if width <= 0 || height <= 0 {
		panic("invalid rectangle for OneBitImage")
	}
	pix := bitmask.New(width * height)
	return &OneBitImage{
		Pix:       pix,
		Stride:    width,
		Rect:      r,
		Threshold: threshold,
	}
}

type OneBitImage struct {
	Pix       *bitmask.Bitmask
	Stride    int
	Rect      image.Rectangle
	Threshold float64
}

func (t OneBitImage) ColorModel() color.Model {
	return OneBitColorModel{Threshold: t.Threshold}
}

func (t OneBitImage) Bounds() image.Rectangle {
	return t.Rect
}

func (t OneBitImage) At(x, y int) color.Color {
	if (image.Point{x, y}.In(t.Rect)) && t.Pix.Get(t.PixOffset(x, y)) {
		return HighColor
	}
	return LowColor
}

func (t OneBitImage) Get(x, y int) bool {
	if !(image.Point{x, y}.In(t.Rect)) {
		return false
	}
	return t.Pix.Get(t.PixOffset(x, y))
}

func (t *OneBitImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(t.Rect)) {
		return
	}
	cm := t.ColorModel()
	if cm.Convert(c) == HighColor {
		t.Pix.Set(t.PixOffset(x, y))
	} else {
		t.Pix.Clear(t.PixOffset(x, y))
	}
}

func (t OneBitImage) PixOffset(x, y int) int {
	return y*t.Stride + x
}

func (t *OneBitImage) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(t.Rect)
	if r.Empty() {
		return &OneBitImage{}
	}
	if r.Eq(t.Rect) {
		return t
	}
	return &OneBitImage{
		Pix:    t.Pix,
		Stride: t.Stride,
		Rect:   r,
	}
}

func (t *OneBitImage) Opaque() bool {
	return true
}

func (t *OneBitImage) Negate() {
	t.Pix.Negate()
}
