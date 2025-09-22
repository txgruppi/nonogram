package printer

import (
	"io"
	"nonogram/board"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var mp *message.Printer

func init() {
	mp = message.NewPrinter(language.English)
}

func PrintBoard(w io.Writer, b *board.Board, count *uint64, started *time.Time) {
	switch {
	case count != nil && started != nil:
		mp.Fprintf(w, "analized %d boards in %s\n", *count, time.Since(*started).String())
	case count != nil && started == nil:
		mp.Fprintf(w, "analized %d boards\n", *count)
	case count == nil && started != nil:
		mp.Fprintf(w, "time taken: %s\n", time.Since(*started).String())
	}
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
	_, _ = w.Write([]byte("\n"))
}

var (
	lastPrint time.Time
)

func PrintBoardOnceEachInterval(interval time.Duration, w io.Writer, b *board.Board, count *uint64, started *time.Time) {
	if time.Since(lastPrint) < interval {
		return
	}
	lastPrint = time.Now()
	PrintBoard(w, b, count, started)
}
