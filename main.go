package main

import (
	"fmt"
	"log"
	"nonogram/encoding"
	"nonogram/image"
	"nonogram/printer"
	"nonogram/solver"
	"os"
)

func run() error {
	file, err := os.Open("board.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	b, hs, err := encoding.Decode(file)
	if err != nil {
		return err
	}

	solved, count, took, err := solver.Solve(b, hs)
	if err != nil {
		return err
	}
	fmt.Printf("analized %d boards in %s\n", count, took.String())
	if solved == nil {
		fmt.Println("no solution found")
		return nil
	}

	printer.PrintBoard(os.Stdout, solved)
	fmt.Println("")
	out, err := os.Create("solution.png")
	if err != nil {
		return err
	}
	defer out.Close()
	err = image.Render(out, solved)
	if err != nil {
		return fmt.Errorf("failed to render image: %w", err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
