package main

import (
	"flag"
	"fmt"
	"go/token"
	"os"
	"path"

	"github.com/ttacon/whodo"
)

var (
	pkg      = flag.String("pkg", "", "package to inspect")
	printNum = flag.Bool("n", false, "print the number of todos per person")
	// TODO(ttacon): add recursive package option (maybe number of
	// package levels to traverse)

	// TODO(ttacon): ensure this exists
	gopath = os.Getenv("GOPATH")
)

func main() {
	flag.Parse()
	if len(*pkg) == 0 {
		flag.Usage()
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

	// TODO(ttacon): do pretty printing (also only show file name since this is
	// per pkg)
	for _, todo := range todos {
		pos := fset.Position(todo.Pos)
		fmt.Printf("%s %s %d %q\n", todo.Author, pos.Filename, pos.Line, todo.Todo)
	}
}
