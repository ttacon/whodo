package main

// TODO(ttacon): open API and move cli to cli package, also add docs

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

var (
	todoMatcher = regexp.MustCompile("TODO\\((.*)\\): (.*)")
	gopath      = os.Getenv("GOPATH")

	pkg      = flag.String("pkg", "", "package to inspect")
	printNum = flag.Bool("n", false, "print the number of todos per person")
)

type Todo struct {
	Pos    token.Pos
	Author string
	Todo   string
}

func main() {
	flag.Parse()
	if len(*pkg) == 0 {
		flag.Usage()
		return
	}

	var (
		todos []Todo
		fset  = token.NewFileSet()
	)

	pkgs, err := parser.ParseDir(fset, path.Join(gopath, "src", *pkg), func(f os.FileInfo) bool {
		return !f.IsDir() && !strings.HasPrefix(f.Name(), ".") && strings.HasSuffix(f.Name(), ".go")
	}, parser.ParseComments)

	if err != nil {
		Log("failed to parse pkg %q, err %v\n", pkg, err)
		return
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, commGroup := range file.Comments {
				for _, comm := range commGroup.List {
					matches := todoMatcher.FindStringSubmatch(comm.Text)
					if len(matches) > 0 {
						todos = append(todos, Todo{
							Pos:    comm.Slash,
							Author: matches[1],
							Todo:   matches[2],
						})
					}
				}
			}
		}
	}

	sort.Sort(byNameAndPosition(todos))

	if *printNum {
		printNumTodos(todos)
		return
	}

	// TODO(ttacon): do pretty printing (also only show file name since this is
	// per pkg)
	for _, todo := range todos {
		pos := fset.Position(todo.Pos)
		fmt.Printf("%s %s %d %q\n", todo.Author, pos.Filename, pos.Line, todo.Todo)
	}

}

func printNumTodos(todos []Todo) {
	var (
		counter     = 0
		lastPerson  = ""
		seen        = make(map[string]int)
		longestName = 0
	)

	for _, todo := range todos {
		if lastPerson != todo.Author {
			if lastPerson == "" {
				counter++
				lastPerson = todo.Author
				longestName = len(lastPerson)
			} else {
				seen[lastPerson] = counter
				counter = 1
				lastPerson = todo.Author
				if len(todo.Author) > longestName {
					longestName = len(todo.Author)
				}
			}
		} else {
			counter++
		}
	}
	seen[lastPerson] = counter
	if len(lastPerson) > longestName {
		longestName = len(lastPerson)
	}

	// pretty print
	// TODO(ttacon): build format string once and use many times
	for name, numTodos := range seen {
		fmt.Printf(fmt.Sprintf("%%%ds %%d\n", longestName), name, numTodos)
	}
}

type byNameAndPosition []Todo

func (b byNameAndPosition) Len() int      { return len(b) }
func (b byNameAndPosition) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byNameAndPosition) Less(i, j int) bool {
	if b[i].Author == b[j].Author {
		return b[i].Pos < b[j].Pos
	}
	return b[i].Author < b[j].Author
}

func Log(message string, args ...interface{}) {
	if message[len(message)-1] != '\n' {
		message += "\n"
	}
	message = "[whodo] " + message
	if !strings.Contains(message, "%") {
		fmt.Println(message)
	}
	fmt.Printf(message, args...)
}
