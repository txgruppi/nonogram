package board

type cellState uint8

const (
	Empty cellState = iota
	Filled
	Crossed
)

func New(width, height int) *Board {
	if width <= 0 || height <= 0 {
		return nil
	}
	return &Board{
		width:  width,
		height: height,
		cells:  make([]cellState, width*height),
	}
}

type Board struct {
	width  int
	height int
	cells  []cellState
}

func (t *Board) index(x, y int) int {
	return y*t.width + x
}

func (t *Board) Size() (width, height int) {
	return t.width, t.height
}

func (t *Board) Get(x, y int) cellState {
	i := t.index(x, y)
	if i < 0 || i >= len(t.cells) {
		return Empty
	}
	return t.cells[i]
}

func (t *Board) Set(x, y int, state cellState) error {
	i := t.index(x, y)
	if i < 0 || i >= len(t.cells) {
		return ErrCellOutOfBounds{}
	}
	if t.cells[i] != Empty {
		return ErrCellAlreadySet{}
	}
	t.cells[i] = state
	return nil
}

func (t *Board) SetRow(y int, states ...cellState) error {
	if y < 0 || y >= t.height || len(states) != t.width {
		return ErrCellOutOfBounds{}
	}
	for x, state := range states {
		if err := t.Set(x, y, state); err != nil {
			return err
		}
	}
	return nil
}

func (t *Board) SetColumn(x int, states ...cellState) error {
	if x < 0 || x >= t.width || len(states) != t.height {
		return ErrCellOutOfBounds{}
	}
	for y, state := range states {
		if err := t.Set(x, y, state); err != nil {
			return err
		}
	}
	return nil
}

func (t *Board) Clone() *Board {
	clone := *t
	clone.cells = make([]cellState, len(t.cells))
	copy(clone.cells, t.cells)
	return &clone
}
