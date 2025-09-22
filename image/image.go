package image

import (
	"image"
	"image/png"
	"io"
	"nonogram/board"
)

const (
	scale = 16
)

var (
	filled  []uint8
	crossed []uint8
	empty   []uint8
)

func init() {
	filled = make([]uint8, scale*scale)
	for i := range filled {
		filled[i] = 0x33
	}
	crossed = make([]uint8, scale*scale)
	for i := range crossed {
		crossed[i] = 0x66
	}
	empty = make([]uint8, scale*scale)
	for i := range empty {
		empty[i] = 0xff
	}
}

func Render(w io.Writer, b *board.Board) error {
	bw, bh := b.Size()
	if bw <= 0 || bh <= 0 {
		return ErrInvalidSize{}
	}
	iw := bw*scale + (bw - 1)
	ih := bh*scale + (bh - 1)
	img := image.NewGray(image.Rect(0, 0, iw, ih))
	for by := 0; by < bh; by++ {
		for bx := 0; bx < bw; bx++ {
			for iy := by*scale + by; iy < (by+1)*scale+by; iy++ {
				for ix := bx * scale; ix < (bx+1)*scale; ix++ {
					switch b.Get(bx, by) {
					case board.Filled:
						copy(img.Pix[iy*iw+ix+bx:], filled[ix%scale:ix%scale+1])
					case board.Crossed:
						copy(img.Pix[iy*iw+ix+bx:], crossed[ix%scale:ix%scale+1])
					case board.Empty:
						copy(img.Pix[iy*iw+ix+bx:], empty[ix%scale:ix%scale+1])
					}
				}
			}
		}
	}
	return png.Encode(w, img)
}
