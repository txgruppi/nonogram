package solver

import (
	"errors"
	"nonogram/board"
	"nonogram/hint"
	"time"
)

func Solve(b *board.Board, hs *hint.Hints) ([]*board.Board, uint64, time.Duration, error) {
	start := time.Now()
	var count uint64
	res, err := solve(b, hs, &count)
	return res, count, time.Since(start), err
}

func solve(b *board.Board, hs *hint.Hints, count *uint64) ([]*board.Board, error) {
	x, y := nextEmpty(b)
	if y == -1 || x == -1 {
		return nil, nil
	}
	solved := make([]*board.Board, 0)

	c := b.Clone()
	if err := c.Set(x, y, board.Filled); err != nil {
		return nil, err
	}
	res, err := try(c, hs, count)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		solved = append(solved, res...)
	}

	c = b.Clone()
	if err := c.Set(x, y, board.Crossed); err != nil {
		return nil, err
	}
	res, err = try(c, hs, count)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		solved = append(solved, res...)
	}

	return solved, nil
}

func try(b *board.Board, hs *hint.Hints, count *uint64) ([]*board.Board, error) {
	*count++
	solved := make([]*board.Board, 0)
	done, err := hs.Check(b)
	if errors.As(err, &hint.ErrInvalid{}) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !done {
		res, err := solve(b, hs, count)
		if err != nil {
			return nil, err
		}
		if len(res) > 0 {
			solved = append(solved, res...)
		}
		return solved, nil
	}
	if done {
		solved = append(solved, b)
	}
	return solved, nil
}

func nextEmpty(b *board.Board) (x, y int) {
	width, height := b.Size()
	for y = 0; y < height; y++ {
		for x = 0; x < width; x++ {
			if b.Get(x, y) == board.Empty {
				return x, y
			}
		}
	}
	return -1, -1
}
