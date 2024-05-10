package commands

import (
	"fmt"
	"io"
	"os"
)

type Command struct {
	PrintUsage func(w io.Writer)
	Run        func(args []string)
}

var registry = make(map[string]*Command)

func Register(name string, command *Command) {
	_, present := registry[name]
	if present {
		fmt.Fprintf(os.Stderr, "internal error: duplicate command %#v\n", name)
		os.Exit(2)
	}
	registry[name] = command
}

func PrintUsage(w io.Writer) {
	fmt.Fprint(w, "Usage:\n")
	for name := range registry {
		registry[name].PrintUsage(w)
	}
}

func Get(name string) *Command {
	return registry[name]
}
