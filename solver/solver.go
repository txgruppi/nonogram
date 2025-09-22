package solver

import (
	"context"
	"errors"
	"nonogram/board"
	"nonogram/hint"
	"nonogram/printer"
	"os"
	"runtime"
	"slices"
	"time"
)

type score struct {
	index           int
	knownCellCount  int
	filledCellCount int
}

type stats struct {
	Start time.Time
	Count uint64
}

func Solve(ctx context.Context, b *board.Board, h *hint.Hints) (*board.Board, stats, error) {
	stats := stats{Start: time.Now()}

	vOrder, hOrder := solveOrder(h)
	solved, err := solve(ctx, b, h, vOrder, hOrder, &stats)
	if err != nil {
		return nil, stats, err
	}
	return solved, stats, nil
}

func solve(ctx context.Context, b *board.Board, h *hint.Hints, vOrder, hOrder []score, stats *stats) (*board.Board, error) {
	runtime.Gosched()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	printer.PrintBoardOnceEachInterval(3*time.Second, os.Stderr, b, &stats.Count, &stats.Start)

	x, y := nextEmpty(b, vOrder, hOrder)
	if x < 0 || y < 0 {
		return nil, nil
	}
	if b.Get(x, y) != board.Empty {
		return nil, nil
	}
	c := b.Clone()
	err := c.Set(x, y, board.Filled)
	if err != nil {
		return nil, err
	}
	solved, err := check(ctx, c, h, vOrder, hOrder, stats)
	if err != nil {
		return nil, err
	}
	if solved != nil {
		return solved, nil
	}
	err = c.Set(x, y, board.Crossed)
	if err != nil {
		return nil, err
	}
	solved, err = check(ctx, c, h, vOrder, hOrder, stats)
	if err != nil {
		return nil, err
	}
	if solved != nil {
		return solved, nil
	}
	return nil, nil
}

func check(ctx context.Context, b *board.Board, h *hint.Hints, vOrder, hOrder []score, stats *stats) (*board.Board, error) {
	stats.Count++

	done, err := h.Check(b)
	if errors.As(err, &hint.ErrInvalid{}) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if done {
		return b, nil
	}
	solved, err := solve(ctx, b, h, vOrder, hOrder, stats)
	if err != nil {
		return nil, err
	}
	if solved != nil {
		return solved, nil
	}
	return nil, nil
}

func scoreSorter(a, b score) int {
	if a.knownCellCount != b.knownCellCount {
		return b.knownCellCount - a.knownCellCount
	}
	if a.filledCellCount != b.filledCellCount {
		return b.filledCellCount - a.filledCellCount
	}
	return a.index - b.index
}

func solveOrder(h *hint.Hints) (vertical, horizontal []score) {
	vertical = make([]score, len(h.Vertical))
	horizontal = make([]score, len(h.Horizontal))

	for i, v := range h.Vertical {
		vertical[i] = score{
			index:           i,
			knownCellCount:  len(v) - 1 + sum(v),
			filledCellCount: sum(v),
		}
	}
	for i, v := range h.Horizontal {
		horizontal[i] = score{
			index:           i,
			knownCellCount:  len(v) - 1 + sum(v),
			filledCellCount: sum(v),
		}
	}

	slices.SortFunc(vertical, scoreSorter)
	slices.SortFunc(horizontal, scoreSorter)

	return vertical, horizontal
}

func nextEmpty(b *board.Board, vs, hs []score) (x, y int) {
	var bestV score
	var bestH score
	var bestVY int
	var bestHX int

findBestVLoop:
	for _, v := range vs {
		for y := 0; y < len(hs); y++ {
			if b.Get(v.index, y) == board.Empty {
				bestV = v
				bestVY = y
				break findBestVLoop
			}
		}
	}

findBestHLoop:
	for _, h := range hs {
		for x := 0; x < len(vs); x++ {
			if b.Get(x, h.index) == board.Empty {
				bestH = h
				bestHX = x
				break findBestHLoop
			}
		}
	}

	if bestV.knownCellCount > bestH.knownCellCount {
		return bestV.index, bestVY
	}
	if bestV.filledCellCount > bestH.filledCellCount {
		return bestV.index, bestVY
	}
	return bestHX, bestH.index
}

func sum(vs []int) int {
	total := 0
	for _, v := range vs {
		total += v
	}
	return total
}
