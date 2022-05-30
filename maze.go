package main

import (
	"github.com/pocke/go-minisat"
	"os"
	"strconv"
	"strings"
)

type Maze struct {
	name    string
	size    int
	data    [][]bool
	solver  *minisat.Solver
	varNum  int
	lineNum int
	cnf     [][]int
}

func NewMaze(path string) *Maze {
	name := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
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
	return &Maze{
		name:    name,
		size:    size,
		data:    data,
		solver:  minisat.NewSolver(0),
		varNum:  0,
		lineNum: 0,
		cnf:     make([][]int, 0),
	}
}

func (maze *Maze) Solve() {
	// TODO: solve the problem
}

func (maze *Maze) OutputCNF(path string) {
	if err := os.Mkdir(path, 0777); err != nil {
		if !os.IsExist(err) {
			panic(err)
		}
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	path += maze.name + ".cnf"
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// TODO: write cnf
}

func (maze *Maze) newVar() *minisat.Var {
	v := maze.solver.NewVar()
	maze.varNum++
	return v
}

func (maze *Maze) addClause(vars ...*minisat.Var) {
	maze.solver.AddClause(vars...)
	maze.cnf = append(maze.cnf, make([]int, len(vars)))
	for i, v := range vars {
		lit := int(*v.CVar) + 1
		if int(*v.CLit)%2 == 1 {
			lit *= -1
		}
		maze.cnf[len(maze.cnf)-1][i] = lit
	}
	maze.lineNum++
}
