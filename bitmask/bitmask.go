package bitmask

func New(n int) *Bitmask {
	if n < 0 {
		n = 0
	}
	return &Bitmask{vs: make([]uint8, (n+7)/8)}
}

type Bitmask struct {
	vs []uint8
}

func (t *Bitmask) index(index int) (byteIndex, bitIndex int) {
	return index / 8, 7 - (index % 8)
}

func (t *Bitmask) Set(index int) {
	byteIndex, bitIndex := t.index(index)
	if byteIndex >= len(t.vs) {
		next := make([]uint8, byteIndex+1)
		copy(next, t.vs)
		t.vs = next
	}
	t.vs[byteIndex] |= 1 << bitIndex
}

func (t *Bitmask) Get(index int) bool {
	byteIndex, bitIndex := t.index(index)
	if byteIndex >= len(t.vs) {
		return false
	}
	return t.vs[byteIndex]&(1<<bitIndex) != 0
}

func (t *Bitmask) Clear(index int) {
	if !t.Get(index) {
		return
	}
	byteIndex, bitIndex := t.index(index)
	t.vs[byteIndex] &^= 1 << bitIndex
}

func (t *Bitmask) Reset() {
	for i := range t.vs {
		t.vs[i] = 0
	}
}

func (t *Bitmask) Grow(highIndex int) {
	byteIndex, _ := t.index(highIndex)
	if byteIndex >= len(t.vs) {
		next := make([]uint8, byteIndex+1)
		copy(next, t.vs)
		t.vs = next
	}
}

func (t Bitmask) Len() int {
	return len(t.vs) * 8
}

func (t *Bitmask) Negate() {
	for i := range t.vs {
		t.vs[i] = ^t.vs[i]
	}
}
