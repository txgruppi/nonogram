package image

type ErrInvalidSize struct {
}

func (e ErrInvalidSize) Error() string {
	return "invalid size"
}

type ErrUnexpectedEmptyCell struct {
}

func (e ErrUnexpectedEmptyCell) Error() string {
	return "unexpected empty cell"
}
