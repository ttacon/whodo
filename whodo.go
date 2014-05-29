package whodo

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	todoMatcher = regexp.MustCompile("TODO\\((.*)\\): (.*)")
	gopath      = os.Getenv("GOPATH")
)

// Todo represents a todo of the form
//     // TODO(ttacon): here's a message
// It holds who the author is, what the message is and
// the position of the todo in a given token.FileSet
type Todo struct {
	Pos    token.Pos
	Author string
	Todo   string
}

// TodosIn returns all of the todos within a given package and
// their positions are with respect to the given fset. If position
// info is not desired, then a token.FileSet doesn't have to be
// passed in. For example:
//
//     fset := token.NewFileSet()
//     todos, err := whodo.TodosIn(fset, "github.com/ttacon/whodo")
//
//     // The above is equivalent to the code below if you
//     // don't care about position info.
//
//     todos, err := whodo.TodosIn(nil, "github.com/ttacon/whodo")
// The returned []Todo is sorted by author name, file in package
// and line in file.
func TodosIn(fset *token.FileSet, pkgPath string) ([]Todo, error) {
	var todos []Todo
	if fset == nil {
		// they didn't feel like providing their own FileSet
		fset = token.NewFileSet()
	}

	pkgs, err := parser.ParseDir(fset, pkgPath, func(f os.FileInfo) bool {
		return !f.IsDir() && !strings.HasPrefix(f.Name(), ".") && strings.HasSuffix(f.Name(), ".go")
	}, parser.ParseComments)

	if err != nil {
		Log("failed to parse pkg %q, err %v\n", pkgPath, err)
		return nil, err
	}

	// need to find a better way...
	currTodo := -1
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, commGroup := range file.Comments {
				inTodo := false
				for _, comm := range commGroup.List {
					matches := todoMatcher.FindStringSubmatch(comm.Text)
					if len(matches) > 0 {
						inTodo = true
						todos = append(todos, Todo{
							Pos:    comm.Slash,
							Author: matches[1],
							Todo:   matches[2],
						})
						currTodo += 1
					} else if inTodo {
						// NOTE(ttacon): maybe only do this if in paren?
						// we already have a todo built, append this comment line to the last todo
						todos[currTodo].Todo = todos[currTodo].Todo + strings.TrimLeft(comm.Text, "//")
					}
				}
			}
		}
	}

	sort.Sort(byNameAndPosition(todos))
	return todos, nil
}

// PrintNumTodos prints each author in the todo slice along with
// the number of todos they have left). Sample output might look
// like:
//
//  ltacon 1
//  ttacon 3
//   b1lly 1000
//
func PrintNumTodos(todos []Todo) {
	var (
		counter     = 0
		longestName = 0
		lastPerson  = ""
		seen        = make(map[string]int)
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

	formatString := fmt.Sprintf("%%%ds %%d\n", longestName)
	for name, numTodos := range seen {
		fmt.Printf(formatString, name, numTodos)
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

// Log logs any messages with the prefix "[whodo]".
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
