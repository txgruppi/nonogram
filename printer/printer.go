package printer

import (
	"io"
	"nonogram/board"
)

func PrintBoard(w io.Writer, b *board.Board) {
	width, height := b.Size()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			switch b.Get(x, y) {
			case board.Empty:
				_, _ = w.Write([]byte("."))
			case board.Filled:
				_, _ = w.Write([]byte("█"))
			case board.Crossed:
				_, _ = w.Write([]byte("X"))
			}
		}
		_, _ = w.Write([]byte("\n"))
	}
}
