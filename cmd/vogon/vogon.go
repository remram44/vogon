package main

import (
	"fmt"
	"log/slog"
	"os"

	_ "github.com/remram44/vogon/internal/apiserver"
	_ "github.com/remram44/vogon/internal/client"
	"github.com/remram44/vogon/internal/commands"
	"github.com/remram44/vogon/internal/versioning"
)

func main() {
	logLevel := slog.LevelInfo

	logLevelStr := os.Getenv("VOGON_LOG_LEVEL")
	if logLevelStr != "" {
		err := logLevel.UnmarshalText([]byte(logLevelStr))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid log level, check $VOGON_LOG_LEVEL\n")
			os.Exit(1)
		}
	}

	logOpts := slog.HandlerOptions{
		Level: logLevel,
	}

	if os.Getenv("VOGON_LOG_JSON") != "" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &logOpts)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &logOpts)))
	}

	slog.Debug("Starting vogon", "logLevel", logLevel, "version", versioning.Version)

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
