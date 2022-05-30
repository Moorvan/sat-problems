package main

import (
	"os"
	"strconv"
	"strings"
)

type Maze struct {
	size int
	data [][]bool
}

func NewMaze(path string) *Maze {
	buf, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	s := string(buf)
	lines := strings.Split(s, "\n")
	size, err := strconv.Atoi(lines[0])
	if err != nil {
		panic(err)
	}
	if len(lines[1:]) != size {
		panic("invalid Maze: lines count != size")
	}
	data := make([][]bool, size)
	for i, line := range lines[1:] {
		if len(line) != size {
			panic("invalid Maze: line length != size in line " + strconv.Itoa(i))
		}
		data[i] = make([]bool, size)
		for j, c := range line {
			if c == '1' {
				data[i][j] = true
			} else {
				data[i][j] = false
			}
		}
	}
	return &Maze{size, data}
}

func (maze Maze) Solve() {

}
