package bitmask

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitmask(t *testing.T) {
	bm := Bitmask{}
	require.Nil(t, bm.vs)

	expected := uint8(0b00000000)
	for i := 0; i < 8; i++ {
		bm.Set(i)
		expected |= 1 << (7 - i)
		require.Equal(t, expected, bm.vs[0])
	}

	bm.Set(98)
	require.Len(t, bm.Len(), 13)
	require.Equal(t, uint8(32), bm.vs[12])

	bm.Clear(0)
	require.Equal(t, uint8(0b01111111), bm.vs[0])

	bm.Reset()
	require.Len(t, bm.Len(), 13)
	for _, v := range bm.vs {
		require.Equal(t, uint8(0), v)
	}

	require.Equal(t, false, bm.Get(10000))

	bm.Grow(105)
	require.Len(t, bm.Len(), 14)
}
