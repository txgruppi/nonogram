package screen

import (
	"errors"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"nonogram/bitmask"
	"nonogram/board"
	"nonogram/hint"
	"os"
	"slices"
	"strconv"
)

var (
	samples15 [10]*OneBitImage
	samples10 [10]*OneBitImage
	samples5  [10]*OneBitImage

	ErrGridSizeNotSupported = errors.New("grid size not supported")
)

func init() {
	for c := 1; c <= 3; c++ {
		for i := 0; i < 10; i++ {
			filepath := "samples/" + strconv.Itoa(c*5) + "_" + strconv.Itoa(i) + ".png"
			sample, err := loadSample(filepath)
			if errors.Is(err, os.ErrNotExist) {
				log.Printf("warning: sample file %q not found, skipping\n", filepath)
				continue
			}
			switch c {
			case 1:
				samples5[i] = sample
			case 2:
				samples10[i] = sample
			case 3:
				samples15[i] = sample
			}
		}
	}
}

func loadSample(filepath string) (*OneBitImage, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	obi := NewOneBitImage(img.Bounds(), .9)
	draw.Draw(obi, img.Bounds(), img, image.Pt(0, 0), draw.Src)
	return obi, nil
}

func DecodeFile(filepath string) (*board.Board, *hint.Hints, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	return Decode(file)
}

func Decode(r io.Reader) (*board.Board, *hint.Hints, error) {
	_ = os.RemoveAll("last")
	_ = os.MkdirAll("last", 0755)

	original, _, err := image.Decode(r)
	if err != nil {
		return nil, nil, err
	}

	obi := NewOneBitImage(original.Bounds(), .9)
	draw.Draw(obi, obi.Bounds(), original, image.Point{original.Bounds().Min.X, original.Bounds().Min.Y}, draw.Src)

	savePNG("last/00-obi.png", obi)

	game := findGameArea(obi, true)

	savePNG("last/01-game.png", game)

	vertical, horizontal, grid := findGameComponents(game)

	savePNG("last/02-vertical.png", vertical)
	savePNG("last/02-horizontal.png", horizontal)
	savePNG("last/02-grid.png", grid)

	verticalCells := findHintCells(vertical)
	for i, cell := range verticalCells {
		savePNG("last/03-vertical-cell-"+strconv.Itoa(i)+".png", cell)
	}

	horizontalCells := findHintCells(horizontal)
	for i, cell := range horizontalCells {
		savePNG("last/03-horizontal-cell-"+strconv.Itoa(i)+".png", cell)
	}

	verticalHits := [][]int{}
	horizontalHits := [][]int{}
	for c, cell := range verticalCells {
		digits := split(cell, all, true)
		for d, digit := range digits {
			savePNG("last/04-vertical-cell-"+strconv.Itoa(c)+"-digit-"+strconv.Itoa(d)+".png", digit)
		}
		numbers, err := identifyNumbers(len(verticalCells), digits)
		if err != nil {
			return nil, nil, err
		}
		verticalHits = append(verticalHits, numbers)
	}
	for c, cell := range horizontalCells {
		digits := split(cell, all, true)
		for d, digit := range digits {
			savePNG("last/04-horizontal-cell-"+strconv.Itoa(c)+"-digit-"+strconv.Itoa(d)+".png", digit)
		}
		numbers, err := identifyNumbers(len(horizontalCells), digits)
		if err != nil {
			return nil, nil, err
		}
		horizontalHits = append(horizontalHits, numbers)
	}

	hints := hint.New(verticalHits, horizontalHits)
	board := readBoard(grid, len(verticalHits), len(horizontalHits))

	os.WriteFile("last/98-hints.txt", []byte(hints.String()), 0644)

	return board, hints, nil
}

func readBoard(obi *OneBitImage, vcount, hcount int) *board.Board {
	b := board.New(vcount, hcount)
	left, top := obi.Bounds().Min.X, obi.Bounds().Min.Y
	width, height := obi.Bounds().Dx(), obi.Bounds().Dy()
	svcount, shcount := vcount/5, hcount/5
	swidth, sheight := width/svcount, height/shcount
	subs := []*OneBitImage{}
	for y := 0; y < shcount; y++ {
		for x := 0; x < svcount; x++ {
			subs = append(subs, obi.SubImage(image.Rect(left+x*swidth, top+y*sheight, left+(x+1)*swidth, top+(y+1)*sheight)).(*OneBitImage))
		}
	}
	for i, sub := range subs {
		left, top := sub.Bounds().Min.X, sub.Bounds().Min.Y
		cellWidth, cellHeight := swidth/5, sheight/5
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				centerX, centerY, leftX := left+x*cellWidth+cellWidth/2, top+y*cellHeight+cellHeight/2, left+x*cellWidth+cellWidth/5
				center, left := !sub.Get(centerX, centerY), !sub.Get(leftX, centerY)
				boardX, boardY := x+(i%svcount)*5, y+(i/svcount)*5
				switch {
				case center && left:
					b.Set(boardX, boardY, board.Filled)

				case center && !left:
					b.Set(boardX, boardY, board.Crossed)
				}
			}
		}
	}
	return b
}

func identifyNumbers(n int, digits []*OneBitImage) ([]int, error) {
	slices.SortStableFunc(digits, func(a, b *OneBitImage) int {
		amin, bmin := a.Bounds().Min, b.Bounds().Min
		ydiff := abs(amin.Y - bmin.Y)
		if ydiff < 5 {
			return amin.X - bmin.X
		}
		return amin.Y - bmin.Y
	})
	numbers := []int{}
	for _, digit := range digits {
		number, err := identifyNumber(n, digit)
		if err != nil {
			return nil, err
		}
		numbers = append(numbers, number)
	}
	var limit int
	switch n {
	case 15:
		limit = 7
	case 10:
		limit = 8
	default:
		return nil, ErrGridSizeNotSupported
	}
	for i := 1; i < len(numbers); i++ {
		aright := digits[i-1].Bounds().Max.X
		bleft := digits[i].Bounds().Min.X
		diff := abs(bleft - aright)
		if diff <= limit {
			numbers[i-1] = numbers[i-1]*10 + numbers[i]
			numbers = append(numbers[:i], numbers[i+1:]...)
		}
	}
	return numbers, nil
}

func identifyNumber(n int, obi *OneBitImage) (int, error) {
	var samples [10]*OneBitImage
	switch n {
	case 5:
		samples = samples5
	case 10:
		samples = samples10
	case 15:
		samples = samples15
	default:
		return -1, ErrGridSizeNotSupported
	}
	left, top := obi.Bounds().Min.X, obi.Bounds().Min.Y
	ow, oh := obi.Bounds().Dx(), obi.Bounds().Dy()
	lowestDiff := math.MaxInt
	lowestIndex := -1
	for i, s := range samples {
		if s == nil {
			continue
		}
		sw, sh := s.Bounds().Dx(), s.Bounds().Dy()
		w := max(ow, sw)
		h := max(oh, sh)
		diff := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if obi.Get(left+x, top+y) != s.Get(x, y) {
					diff++
				}
			}
		}
		if diff < lowestDiff {
			lowestDiff = diff
			lowestIndex = i
		}
	}
	return lowestIndex, nil
}

func findHintCells(obi *OneBitImage) []*OneBitImage {
	cells := split(obi, all, true)
	for i := range cells {
		cells[i] = removeBorder(cells[i], false)
		cells[i] = shrinkWhile(cells[i], all, true)
	}
	return cells
}

func findGameArea(obi *OneBitImage, horizontalCuts bool) *OneBitImage {
	width, height := obi.Bounds().Max.X, obi.Bounds().Max.Y
	sections := [][2]int{}
	start := 0
	end := 0
	if horizontalCuts {
		for {
			if all(obi, image.Rect(0, start, width, start+1), true) {
				start++
				continue
			}
			end = start
			for !all(obi, image.Rect(0, end, width, end+1), true) {
				end++
				if end >= height {
					break
				}
			}
			if end >= height {
				break
			}
			sections = append(sections, [2]int{start, end})
			start = end
		}
	}
	return shrinkWhile(obi.SubImage(image.Rect(0, sections[3][0], obi.Bounds().Dx(), sections[4][1])).(*OneBitImage), all, true)
}

func findGameComponents(obi *OneBitImage) (*OneBitImage, *OneBitImage, *OneBitImage) {
	left, top, right, bottom := obi.Bounds().Min.X, obi.Bounds().Min.Y, obi.Bounds().Max.X, obi.Bounds().Max.Y
	for some(obi, image.Rect(left, top, right, top+1), false) {
		top++
	}
	for some(obi, image.Rect(left, top, left+1, bottom), false) {
		left++
	}
	vertical := shrinkWhile(obi.SubImage(image.Rect(left, obi.Rect.Min.Y, right, top)).(*OneBitImage), all, true)
	horizontal := shrinkWhile(obi.SubImage(image.Rect(obi.Rect.Min.X, top, left, bottom)).(*OneBitImage), all, true)
	grid := shrinkWhile(obi.SubImage(image.Rect(left, top, right, bottom)).(*OneBitImage), all, true)
	return vertical, horizontal, grid
}

func all(obi *OneBitImage, r image.Rectangle, value bool) bool {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if obi.Get(x, y) != value {
				return false
			}
		}
	}
	return true
}

func some(obi *OneBitImage, r image.Rectangle, value bool) bool {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if obi.Get(x, y) == value {
				return true
			}
		}
	}
	return false
}

func shrinkWhile(obi *OneBitImage, cond func(*OneBitImage, image.Rectangle, bool) bool, value bool) *OneBitImage {
	left, top, right, bottom := obi.Bounds().Min.X, obi.Bounds().Min.Y, obi.Bounds().Max.X, obi.Bounds().Max.Y
	for cond(obi, image.Rect(left, top, right, top+1), value) {
		top++
	}
	for cond(obi, image.Rect(left, bottom-1, right, bottom), value) {
		bottom--
	}
	for cond(obi, image.Rect(left, top, left+1, bottom), value) {
		left++
	}
	for cond(obi, image.Rect(right-1, top, right, bottom), value) {
		right--
	}
	if left >= right || top >= bottom {
		return nil
	}
	return obi.SubImage(image.Rect(left, top, right, bottom)).(*OneBitImage)
}

func split(obi *OneBitImage, cond func(*OneBitImage, image.Rectangle, bool) bool, value bool) []*OneBitImage {
	res := []*OneBitImage{}
	stack := []*OneBitImage{obi}
	sp := 0

	for sp < len(stack) {
		curr := stack[sp]
		sp++

		width, height := curr.Bounds().Dx(), curr.Bounds().Dy()
		verticalChecks := bitmask.New(width)
		horizontalChecks := bitmask.New(height)

		left, top, right, bottom := curr.Bounds().Min.X, curr.Bounds().Min.Y, curr.Bounds().Max.X, curr.Bounds().Max.Y
		for x := left; x < right; x++ {
			if cond(curr, image.Rect(x, top, x+1, bottom), value) {
				verticalChecks.Set(x - left)
			}
		}
		for y := top; y < bottom; y++ {
			if cond(curr, image.Rect(left, y, right, y+1), value) {
				horizontalChecks.Set(y - top)
			}
		}

		verticalChanges := []int{}
		for x := 1; x < width && x < verticalChecks.Len(); x++ {
			if verticalChecks.Get(x) != verticalChecks.Get(x-1) {
				verticalChanges = append(verticalChanges, x)
			}
		}

		if len(verticalChanges) > 0 {
			verticalChanges = append(verticalChanges, right-left)
			verticalChanges = append([]int{0}, verticalChanges...)
			inAreaToSplit := cond(curr, image.Rect(left, top, left+1, bottom), value)
			for i := 1; i < len(verticalChanges); i++ {
				if !inAreaToSplit {
					stack = append(stack, curr.SubImage(image.Rect(left+verticalChanges[i-1], top, left+verticalChanges[i], bottom)).(*OneBitImage))
				}
				inAreaToSplit = !inAreaToSplit
			}
			continue
		}

		horizontalChanges := []int{}
		for y := 1; y < height && y < horizontalChecks.Len(); y++ {
			if horizontalChecks.Get(y) != horizontalChecks.Get(y-1) {
				horizontalChanges = append(horizontalChanges, y)
			}
		}

		if len(horizontalChanges) > 0 {
			horizontalChanges = append(horizontalChanges, bottom-top)
			horizontalChanges = append([]int{0}, horizontalChanges...)
			inAreaToSplit := cond(curr, image.Rect(left, top, right, top+1), value)
			for i := 1; i < len(horizontalChanges); i++ {
				if !inAreaToSplit {
					stack = append(stack, curr.SubImage(image.Rect(left, top+horizontalChanges[i-1], right, top+horizontalChanges[i])).(*OneBitImage))
				}
				inAreaToSplit = !inAreaToSplit
			}
			continue
		}

		res = append(res, curr)
	}

	return res
}

func removeBorder(obi *OneBitImage, value bool) *OneBitImage {
	width, height := obi.Bounds().Dx(), obi.Bounds().Dy()
	left, top, right, bottom := obi.Bounds().Min.X, obi.Bounds().Min.Y, obi.Bounds().Max.X, obi.Bounds().Max.Y
	centerX, centerY := width/2+left, height/2+top
	for obi.Get(centerX, top) == value {
		top++
	}
	for obi.Get(centerX, bottom-1) == value {
		bottom--
	}
	for obi.Get(left, centerY) == value {
		left++
	}
	for obi.Get(right-1, centerY) == value {
		right--
	}
	return shrinkWhile(obi.SubImage(image.Rect(left, top, right, bottom)).(*OneBitImage), some, false)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func savePNG(filepath string, img image.Image) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}
