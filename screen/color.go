package screen

var (
	HighColor = OneBitColor{high: true}
	LowColor  = OneBitColor{high: false}
)

type OneBitColor struct {
	high bool
}

func (c OneBitColor) RGBA() (r, g, b, a uint32) {
	if c.high {
		return 0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff
	}
	return 0x00000000, 0x00000000, 0x00000000, 0xffffffff
}
