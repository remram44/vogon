package main

import (
	"fmt"
	"os"

	_ "github.com/remram44/vogon/internal/apiserver"
	_ "github.com/remram44/vogon/internal/client"
	"github.com/remram44/vogon/internal/commands"
)

func main() {
	usage := func(code int) {
		commands.PrintUsage(os.Stderr)
		os.Exit(code)
	}

	if len(os.Args) < 2 {
		usage(2)
	}
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "help", "-help", "--help":
			usage(0)
		}
	}

	command := commands.Get(os.Args[1])
	if command == nil {
		fmt.Fprintf(os.Stderr, "No such command: %v\n\n", os.Args[1])
		usage(1)
	}

	err := command.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
