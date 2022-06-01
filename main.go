package main

import (
	"io/fs"
	"path/filepath"
)

func main() {
	if err := filepath.WalkDir("./cases", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		maze := NewMaze(path)
		maze.Solve(100)
		return nil
	}); err != nil {
		panic(err)
	}
}
