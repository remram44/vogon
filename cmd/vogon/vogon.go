package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	_ "github.com/remram44/vogon/internal/apiserver"
	_ "github.com/remram44/vogon/internal/client"
	"github.com/remram44/vogon/internal/commands"
)

func main() {
	logLevelStr := os.Getenv("VOGON_LOG_LEVEL")
	if logLevelStr != "" {
		logLevel, err := log.ParseLevel(logLevelStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid log level, check $VOGON_LOG_LEVEL\n")
			os.Exit(1)
		}
		log.SetLevel(logLevel)
		log.Infof("Set log level to %v", logLevel)
	}

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
