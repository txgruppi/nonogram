package hint

import (
	"nonogram/board"
)

func New(vertical, horizontal [][]int) *Hints {
	return &Hints{
		Vertical:   vertical,
		Horizontal: horizontal,
	}
}

type Hints struct {
	Vertical   [][]int
	Horizontal [][]int
}

func (t *Hints) Check(b *board.Board) (bool, error) {
	width, height := b.Size()
	if len(t.Vertical) != width || len(t.Horizontal) != height {
		return false, ErrInvalidBoardSize{
			expected: [2]int{len(t.Vertical), len(t.Horizontal)},
			actual:   [2]int{width, height},
		}
	}
	hasEmpty := false
	for y, hint := range t.Horizontal {
		if len(hint) == 0 {
			return false, ErrMissingHints{direction: "horizontal", index: y}
		}
		empty, filled := t.runs(b, 0, y, width, y+1)
		if len(empty) == 0 {
			if len(filled) != len(hint) {
				return false, ErrInvalid{isRow: true, index: y}
			}
			for i, count := range filled {
				if count != hint[i] {
					return false, ErrInvalid{isRow: true, index: y}
				}
			}
			continue
		}

		if len(filled) > len(hint) {
			return false, ErrInvalid{isRow: true, index: y}
		}

		totalAvailable := sum(empty) + sum(filled)
		totalHint := sum(hint)
		if totalAvailable < totalHint {
			return false, ErrInvalid{isRow: true, index: y}
		}

		for i := 0; i < len(filled); i++ {
			if i < len(hint) && filled[i] > hint[i] {
				return false, ErrInvalid{isRow: true, index: y}
			}
		}

		hasEmpty = true
	}
	for x, hint := range t.Vertical {
		if len(hint) == 0 {
			return false, ErrMissingHints{direction: "vertical", index: x}
		}
		empty, filled := t.runs(b, x, 0, x+1, height)
		if len(empty) == 0 {
			if len(filled) != len(hint) {
				return false, ErrInvalid{isRow: false, index: x}
			}
			for i, count := range filled {
				if count != hint[i] {
					return false, ErrInvalid{isRow: false, index: x}
				}
			}
			continue
		}

		if len(filled) > len(hint) {
			return false, ErrInvalid{isRow: false, index: x}
		}

		totalAvailable := sum(empty) + sum(filled)
		totalHint := sum(hint)
		if totalAvailable < totalHint {
			return false, ErrInvalid{isRow: false, index: x}
		}

		for i := 0; i < len(filled); i++ {
			if i < len(hint) && filled[i] > hint[i] {
				return false, ErrInvalid{isRow: false, index: x}
			}
		}

		hasEmpty = true
	}
	return !hasEmpty, nil
}

func (t *Hints) runs(b *board.Board, x1, y1, x2, y2 int) (empty, filled []int) {
	if x2-x1 > 1 && y2-y1 > 1 {
		return nil, nil
	}
	first := true
	last := board.Empty
	count := 0
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			cell := b.Get(x, y)
			if first {
				first = false
				last = cell
				count = 1
				continue
			}
			if cell == last {
				count++
				continue
			}
			switch last {
			case board.Empty:
				empty = append(empty, count)
			case board.Filled:
				filled = append(filled, count)
			case board.Crossed:
			}
			last = cell
			count = 1
		}
	}
	if count > 0 {
		switch last {
		case board.Empty:
			empty = append(empty, count)
		case board.Filled:
			filled = append(filled, count)
		case board.Crossed:
		}
	}
	return empty, filled
}

func sum(s []int) int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

func max(s []int) int {
	if len(s) == 0 {
		return 0
	}
	v := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] > v {
			v = s[i]
		}
	}
	return v
}
