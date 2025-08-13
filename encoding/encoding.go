package encoding

import (
	"bufio"
	"errors"
	"io"
	"nonogram/board"
	"nonogram/hint"
	"strconv"
	"strings"
)

func Decode(r io.Reader) (*board.Board, *hint.Hints, error) {
	scanner := bufio.NewScanner(r)
	width, height, err := decodeSize(scanner)
	if err != nil {
		return nil, nil, err
	}
	b := board.New(width, height)
	vertical := make([][]int, width)
	horizontal := make([][]int, height)
	for x := 0; x < width; x++ {
		hints, err := decodeHints(scanner)
		if err != nil {
			return nil, nil, err
		}
		vertical[x] = hints
	}
	for y := 0; y < height; y++ {
		hints, err := decodeHints(scanner)
		if err != nil {
			return nil, nil, err
		}
		horizontal[y] = hints
	}
	for {
		x, y, err := decodePosition(scanner)
		if errors.As(err, &ErrUnexpectedEOF{}) {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		b.Set(x, y, board.Crossed)
	}
	return b, hint.New(vertical, horizontal), nil
}

func decodeSize(scanner *bufio.Scanner) (width, height int, err error) {
	if !scanner.Scan() {
		return 0, 0, ErrUnexpectedEOF{}
	}
	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return 0, 0, ErrInvalidSize{}
	}
	width, err = strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, err
	}
	height, err = strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, err
	}
	if width <= 0 || height <= 0 {
		return 0, 0, ErrInvalidSize{}
	}
	return width, height, nil
}

func decodeHints(scanner *bufio.Scanner) ([]int, error) {
	if !scanner.Scan() {
		return nil, ErrUnexpectedEOF{}
	}
	line := scanner.Text()
	fields := strings.Fields(line)
	hints := make([]int, len(fields))
	for i, field := range fields {
		hint, err := strconv.Atoi(field)
		if err != nil {
			return nil, err
		}
		if hint < 0 {
			return nil, ErrInvalidHints{}
		}
		hints[i] = hint
	}
	return hints, nil
}

func decodePosition(scanner *bufio.Scanner) (int, int, error) {
	if !scanner.Scan() {
		return -1, -1, ErrUnexpectedEOF{}
	}
	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return -1, -1, ErrInvalidPosition{}
	}
	x, err := strconv.Atoi(fields[0])
	if err != nil || x < 0 {
		return -1, -1, err
	}
	y, err := strconv.Atoi(fields[1])
	if err != nil || y < 0 {
		return -1, -1, err
	}
	return x, y, nil
}
