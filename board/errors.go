package board

type ErrCellOutOfBounds struct {
}

func (e ErrCellOutOfBounds) Error() string {
	return "cell out of bounds"
}

type ErrCellAlreadySet struct {
}

func (e ErrCellAlreadySet) Error() string {
	return "cell already set"
}
