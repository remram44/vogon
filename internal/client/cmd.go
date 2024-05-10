package client

import (
	"fmt"
	"io"

	"github.com/remram44/vogon/internal/commands"
)

func get(args []string) error {
	// TODO
	return nil
}

func apply(args []string) error {
	// TODO
	return nil
}

func init() {
	commands.Register("get", &commands.Command{
		PrintUsage: func(w io.Writer) {
			fmt.Fprintf(
				w,
				""+
					"  get <name>\n"+
					"    Get an object from the API\n",
			)
		},
		Run: get,
	})
	commands.Register("apply", &commands.Command{
		PrintUsage: func(w io.Writer) {
			fmt.Fprintf(
				w,
				""+
					"  apply\n"+
					"    Create/replace/update object(s) from JSON on stdin\n",
			)
		},
		Run: apply,
	})
}
