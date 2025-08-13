package printer

import (
	"fmt"
	"io"
	"nonogram/board"
	"time"
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

var (
	lastPrint time.Time
)

func PrintBoardOnceASecond(w io.Writer, b *board.Board, count *uint64) {
	if time.Since(lastPrint) < time.Second {
		return
	}
	lastPrint = time.Now()
	if count != nil {
		fmt.Fprintf(w, "%d\n", *count)
	}
	PrintBoard(w, b)
	_, _ = w.Write([]byte("\n"))
}
