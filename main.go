package main

import "github.com/pocke/go-minisat"

func main() {
	s := minisat.NewSolver(0)
	v1 := s.NewVar()
	v2 := s.NewVar()
	s.AddClause(v1, v2)
	s.AddClause(v1, v2.Not())
	if s.Solve() {
		println("satisfiable")
		println(s.ModelValue(v1))
		println(s.ModelValue(v2))
	} else {
		println("unsatisfiable")
	}
}
