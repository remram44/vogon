package commands

import (
	"fmt"
	"io"
	"os"
	"slices"
)

type Command struct {
	PrintUsage func(w io.Writer)
	Run        func(args []string) error
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
	keys := make([]string, 0, len(registry))
	for name := range registry {
		keys = append(keys, name)
	}
	slices.Sort(keys)

	fmt.Fprint(w, "Usage:\n")
	for _, name := range keys {
		registry[name].PrintUsage(w)
	}
}

func Get(name string) *Command {
	return registry[name]
}
