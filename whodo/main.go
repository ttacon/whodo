package main

import (
	"flag"
	"fmt"
	"go/token"
	"os"
	"path"
	"path/filepath"

	"github.com/ttacon/whodo"
)

var (
	pkg      = flag.String("pkg", "", "package to inspect")
	printNum = flag.Bool("n", false, "print the number of todos per person")
	// TODO(ttacon): add recursive package option (maybe number of
	// package levels to traverse)

	gopath = os.Getenv("GOPATH")
)

func main() {
	flag.Parse()
	if len(*pkg) == 0 {
		flag.Usage()
		return
	}

	if len(gopath) == 0 {
		whodo.Log("GOPATH must be set")
		return
	}

	var (
		pkgPath = path.Join(gopath, "src", *pkg)
		fset    = token.NewFileSet()
	)

	todos, err := whodo.TodosIn(fset, pkgPath)
	if err != nil {
		return
	}

	if *printNum {
		whodo.PrintNumTodos(todos)
		return
	}

	printTodos(todos, fset)
}

func printTodos(todos []whodo.Todo, fset *token.FileSet) {
	var (
		authLen, fnameLen int
	)

	for _, todo := range todos {
		if len(todo.Author) > authLen {
			authLen = len(todo.Author)
		}

		if pos := fset.Position(todo.Pos); len(filepath.Base(pos.Filename)) > fnameLen {
			fnameLen = len(filepath.Base(pos.Filename))
		}
	}

	// TODO(ttacon): make printing of line numbers prettier
	fmtString := fmt.Sprintf("%%%ds  %%%ds  %%4d  %%q\n", authLen, fnameLen)
	for _, todo := range todos {
		pos := fset.Position(todo.Pos)
		fmt.Printf(fmtString, todo.Author, filepath.Base(pos.Filename), pos.Line, todo.Todo)
	}
}
