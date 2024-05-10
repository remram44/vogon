package apiserver

import (
	"fmt"
	"io"

	"github.com/remram44/vogon/internal/commands"
)

func PrintUsage(w io.Writer) {
	fmt.Fprintf(
		w,
		""+
			"  apiserver <config>\n"+
			"    Run the API server\n",
	)
}

func Run(args []string) {
	// TODO
	fmt.Printf("apiserver %#v\n", args)
}

func init() {
	commands.Register("apiserver", &commands.Command{
		PrintUsage: PrintUsage,
		Run:        Run,
	})
}
