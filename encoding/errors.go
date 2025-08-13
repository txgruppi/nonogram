package encoding

type ErrUnexpectedEOF struct {
}

func (e ErrUnexpectedEOF) Error() string {
	return "unexpected EOF"
}

type ErrInvalidSize struct {
}

func (e ErrInvalidSize) Error() string {
	return "invalid size"
}

type ErrInvalidHints struct {
}

func (e ErrInvalidHints) Error() string {
	return "invalid hints"
}

type ErrInvalidPosition struct {
}

func (e ErrInvalidPosition) Error() string {
	return "invalid position"
}
