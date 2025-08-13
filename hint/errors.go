package hint

import "fmt"

type ErrInvalidBoardSize struct {
	expected [2]int
	actual   [2]int
}

func (e ErrInvalidBoardSize) Error() string {
	return fmt.Sprintf("invalid board size: expected %dx%d, got %dx%d", e.expected[0], e.expected[1], e.actual[0], e.actual[1])
}

type ErrMissingHints struct {
	direction string
	index     int
}

func (e ErrMissingHints) Error() string {
	return fmt.Sprintf("missing %s hints at index %d", e.direction, e.index)
}

type ErrInvalid struct {
	isRow bool
	index int
}

func (e ErrInvalid) Error() string {
	if e.isRow {
		return fmt.Sprintf("invalid row at index %d", e.index)
	}
	return fmt.Sprintf("invalid column at index %d", e.index)
}
