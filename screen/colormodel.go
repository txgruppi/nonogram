package screen

import "image/color"

type OneBitColorModel struct {
	Threshold float64
}

func (t OneBitColorModel) Convert(c color.Color) color.Color {
	if _, ok := c.(OneBitColor); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	if uint8(a) < 128 {
		return LowColor
	}
	rf := float64(uint8(r)) / 255.0
	gf := float64(uint8(g)) / 255.0
	bf := float64(uint8(b)) / 255.0
	avg := .299*rf + .587*gf + .114*bf
	if avg < t.Threshold {
		return LowColor
	}
	return HighColor
}
