package board

type ErrCellOutOfBounds struct {
}

func (e ErrCellOutOfBounds) Error() string {
	return "cell out of bounds"
}
