package main

import (
	"fmt"
	"github.com/pocke/go-minisat"
	"log"
	"os"
	"strconv"
	"strings"
)

type Maze struct {
	name     string
	size     int
	data     [][]bool
	solver   *minisat.Solver
	varNum   int
	lineNum  int
	cnf      [][]int
	var2name map[*minisat.Var]string
	name2var map[string]*minisat.Var
}

type Domain string

const (
	person  Domain = "person"
	blocked Domain = "blocked"
	empty   Domain = "empty"
)

type Direction string

const (
	left  Direction = "left"
	right Direction = "right"
	up    Direction = "up"
	down  Direction = "down"
)

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
		name:     name,
		size:     size,
		data:     data,
		solver:   minisat.NewSolver(0),
		varNum:   0,
		lineNum:  0,
		cnf:      make([][]int, 0),
		var2name: make(map[*minisat.Var]string),
		name2var: make(map[string]*minisat.Var),
	}
}

func (maze *Maze) Solve(step int) {
	//fmt.Println(maze.data)
	for i := 1; i < step; i++ {
		log.Println("step", i)
		maze.setConstraints(i)
		if maze.solver.Solve() {
			log.Println("Solved in step", i)
			log.Printf("solution written to ./result/%s%s", maze.name, ".model")
			maze.OutputModel("./result", i)
			log.Printf("cnf written to ./cnf/%s%s", maze.name, ".cnf")
			maze.OutputCNF("./cnf")
			return
		} else {
			log.Println("Unsolved in step", i)
		}
	}
	log.Println("Can't find solution in", step, "steps.")
}

func (maze *Maze) setConstraints(k int) {
	maze.cleanSolver()
	maze.setMazeConstraint(k)
	maze.setInit()
	maze.setGoal(k)
	maze.constraint(k)
	maze.setActions(k)
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
	writeLine := func(s string) {
		if _, err := f.WriteString(s + "\n"); err != nil {
			panic(err)
		}
	}
	writeInt := func(i int) {
		if _, err := f.WriteString(strconv.Itoa(i) + " "); err != nil {
			panic(err)
		}
	}
	writeLine("p cnf " + strconv.Itoa(maze.varNum) + " " + strconv.Itoa(len(maze.cnf)))

	for _, clause := range maze.cnf {
		for _, lit := range clause {
			writeInt(lit)
		}
		writeLine("0")
	}
}

func (maze *Maze) OutputModel(path string, k int) {
	if err := os.Mkdir(path, 0777); err != nil {
		if !os.IsExist(err) {
			panic(err)
		}
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	path += maze.name + ".model"
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for t := 0; t < k; t++ {
		for _, pos := range maze.getEmptyPositions() {
			x, y := pos.x, pos.y
			writeMsg := func(d Direction) {
				if _, err := f.WriteString(fmt.Sprintf("@%d: (%d, %d) move %s\n", t, x, y, d)); err != nil {
					panic(err)
				}
			}
			move := func(d Direction) bool {
				a := maze.getAction(x, y, d, t)
				if res, err := maze.solver.ModelValue(a); err != nil {
					panic(err)
				} else if res {
					writeMsg(d)
					return true
				}
				return false
			}
			if move(left) || move(right) || move(up) || move(down) {
				break
			}
		}
	}
}

func (maze *Maze) getVar(name string) *minisat.Var {
	if v, ok := maze.name2var[name]; ok {
		return v
	}
	v := maze.solver.NewVar()
	maze.varNum++
	maze.var2name[v] = name
	maze.name2var[name] = v
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

func (maze *Maze) cleanSolver() {
	maze.solver = minisat.NewSolver(0)
	maze.varNum = 0
	maze.lineNum = 0
	maze.cnf = make([][]int, 0)
	maze.var2name = make(map[*minisat.Var]string)
	maze.name2var = make(map[string]*minisat.Var)
}

func (maze *Maze) getState(x, y int, domain Domain, time int) *minisat.Var {
	name := fmt.Sprintf("state@(%d, %d)=%s@%d", x, y, domain, time)
	return maze.getVar(name)
}

func (maze *Maze) getAction(x, y int, direction Direction, time int) *minisat.Var {
	name := fmt.Sprintf("action@%s@(%d, %d)@%d", direction, x, y, time)
	return maze.getVar(name)
}

func (maze *Maze) setMazeConstraint(k int) {
	for t := 1; t <= k; t++ {
		for x := 0; x < maze.size; x++ {
			for y := 0; y < maze.size; y++ {
				if maze.data[x][y] {
					maze.addClause(maze.getState(x, y, blocked, t))
					maze.addClause(maze.getState(x, y, person, t).Not())
					maze.addClause(maze.getState(x, y, empty, t).Not())
				} else {
					maze.addClause(maze.getState(x, y, blocked, t).Not())
				}
			}
		}
	}
}

func (maze *Maze) setInit() {
	if maze.data[0][0] {
		panic("invalid Maze: start point is blocked")
	}

	for x := 0; x < maze.size; x++ {
		for y := 0; y < maze.size; y++ {
			if x == 0 && y == 0 {
				maze.addClause(maze.getState(x, y, person, 0))
				maze.addClause(maze.getState(x, y, empty, 0).Not())
			} else {
				if !maze.data[x][y] {
					maze.addClause(maze.getState(x, y, empty, 0))
					maze.addClause(maze.getState(x, y, person, 0).Not())
				}
			}
		}
	}
}

func (maze *Maze) setGoal(k int) {
	maze.addClause(maze.getState(maze.size-1, maze.size-1, person, k))
}

func (maze *Maze) constraint(k int) {
	for t := 1; t <= k; t++ {
		for x := 0; x < maze.size; x++ {
			for y := 0; y < maze.size; y++ {
				maze.addClause(maze.getState(x, y, person, t),
					maze.getState(x, y, empty, t),
					maze.getState(x, y, blocked, t))

				maze.addClause(maze.getState(x, y, person, t).Not(),
					maze.getState(x, y, empty, t).Not())
				maze.addClause(maze.getState(x, y, person, t).Not(),
					maze.getState(x, y, blocked, t).Not())
				maze.addClause(maze.getState(x, y, empty, t).Not(),
					maze.getState(x, y, blocked, t).Not())
			}
		}
	}
}

type Position struct {
	x, y int
}

func (maze *Maze) getEmptyPositions() []Position {
	emptyPositions := make([]Position, 0)
	for x := 0; x < maze.size; x++ {
		for y := 0; y < maze.size; y++ {
			if !maze.data[x][y] {
				emptyPositions = append(emptyPositions, Position{x, y})
			}
		}
	}
	return emptyPositions
}

func (maze *Maze) setActions(k int) {
	for t := 0; t < k; t++ {
		actionsAtT := make([]*minisat.Var, 0)
		for _, pos := range maze.getEmptyPositions() {
			x, y := pos.x, pos.y
			leave := make([]*minisat.Var, 0)
			come := make([]*minisat.Var, 0)
			moveLeft := maze.getAction(x, y, left, t)
			moveRight := maze.getAction(x, y, right, t)
			moveUp := maze.getAction(x, y, up, t)
			moveDown := maze.getAction(x, y, down, t)
			if x > 0 && !maze.data[x-1][y] {
				a := moveUp
				leave = append(leave, a)
				come = append(come, maze.getAction(x-1, y, down, t))
				maze.addClause(a.Not(), maze.getState(x, y, person, t))
				maze.addClause(a.Not(), maze.getState(x-1, y, empty, t))
				maze.addClause(a.Not(), maze.getState(x, y, empty, t+1))
				maze.addClause(a.Not(), maze.getState(x-1, y, person, t+1))
			} else {
				maze.addClause(moveUp.Not())
			}
			if x < maze.size-1 && !maze.data[x+1][y] {
				a := moveDown
				leave = append(leave, a)
				come = append(come, maze.getAction(x+1, y, up, t))
				maze.addClause(a.Not(), maze.getState(x, y, person, t))
				maze.addClause(a.Not(), maze.getState(x+1, y, empty, t))
				maze.addClause(a.Not(), maze.getState(x, y, empty, t+1))
				maze.addClause(a.Not(), maze.getState(x+1, y, person, t+1))
			} else {
				maze.addClause(moveDown.Not())
			}
			if y > 0 && !maze.data[x][y-1] {
				a := moveLeft
				leave = append(leave, a)
				come = append(come, maze.getAction(x, y-1, right, t))
				maze.addClause(a.Not(), maze.getState(x, y, person, t))
				maze.addClause(a.Not(), maze.getState(x, y-1, empty, t))
				maze.addClause(a.Not(), maze.getState(x, y, empty, t+1))
				maze.addClause(a.Not(), maze.getState(x, y-1, person, t+1))
			} else {
				maze.addClause(moveLeft.Not())
			}
			if y < maze.size-1 && !maze.data[x][y+1] {
				a := moveRight
				leave = append(leave, a)
				come = append(come, maze.getAction(x, y+1, left, t))
				maze.addClause(a.Not(), maze.getState(x, y, person, t))
				maze.addClause(a.Not(), maze.getState(x, y+1, empty, t))
				maze.addClause(a.Not(), maze.getState(x, y, empty, t+1))
				maze.addClause(a.Not(), maze.getState(x, y+1, person, t+1))
			} else {
				maze.addClause(moveRight.Not())
			}
			actionsAtT = append(actionsAtT, leave...)
			maze.addClause(append(leave,
				maze.getState(x, y, person, t).Not(),
				maze.getState(x, y, empty, t+1).Not())...)
			maze.addClause(append(come,
				maze.getState(x, y, empty, t).Not(),
				maze.getState(x, y, person, t+1).Not())...)
		}
		maze.addClause(actionsAtT...)
		for i, a1 := range actionsAtT {
			for _, a2 := range actionsAtT[i+1:] {
				maze.addClause(a1.Not(), a2.Not())
			}
		}
	}
}

func (maze *Maze) setNoRing(k int) {
	for _, pos := range maze.getEmptyPositions() {
		x, y := pos.x, pos.y
		for i := 0; i <= k; i++ {
			for j := i + 1; j <= k; j++ {
				maze.addClause(maze.getState(x, y, person, i).Not(),
					maze.getState(x, y, person, j).Not())
			}
		}
	}
}
