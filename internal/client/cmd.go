package client

import (
	"fmt"
	"io"
	"os"

	"github.com/remram44/vogon/internal/commands"
	"github.com/remram44/vogon/internal/versioning"
)

func GetClientFromEnv() (*Client, error) {
	options := ClientOptions{
		Uri: os.Getenv("VOGON_SERVER_URI"),
	}
	client, err := NewClient(options)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func version(args []string) error {
	fmt.Printf("Client: %s\n", versioning.NameAndVersionString())

	client, err := GetClientFromEnv()
	if err != nil {
		return err
	}

	version, err := client.GetVersion()
	if err != nil {
		return err
	}

	fmt.Printf("Server: %s\n", version)

	return nil
}

func get(args []string) error {
	// TODO
	return nil
}

func apply(args []string) error {
	// TODO
	return nil
}

func init() {
	commands.Register("version", &commands.Command{
		PrintUsage: func(w io.Writer) {
			fmt.Fprintf(
				w,
				""+
					"  version\n"+
					"    Print client and server versions\n",
			)
		},
		Run: version,
	})
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
