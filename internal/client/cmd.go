package client

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/remram44/vogon/internal/commands"
	"github.com/remram44/vogon/internal/database"
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
	if len(args) != 2 {
		return fmt.Errorf("Missing object name")
	}

	name := args[1]

	client, err := GetClientFromEnv()
	if err != nil {
		return err
	}

	object, err := client.GetObject(name)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(object)

	return nil
}

func apply(args []string) error {
	create := true
	replace := true
	stripRevision := false
	for _, opt := range args[1:] {
		switch opt {
		case "--if-not-exists":
			if !create {
				return fmt.Errorf("Incompatible options")
			}
			replace = false
			stripRevision = true
		case "--force-overwrite":
			if !replace {
				return fmt.Errorf("Incompatible options")
			}
			stripRevision = true
		case "--no-create":
			if !replace {
				return fmt.Errorf("Incompatible options")
			}
			create = false
		default:
			fmt.Fprintf(os.Stderr, "Unknown option: %v", opt)
			os.Exit(2)
		}
	}

	decoder := yaml.NewDecoder(os.Stdin)
	decoder.KnownFields(true)
	var object database.Object
	err := decoder.Decode(&object)
	if err != nil {
		return err
	}

	valid := true
	if object.Kind == "" {
		fmt.Fprintf(os.Stderr, "Missing kind\n")
		valid = false
	}
	if object.Metadata.Name == "" {
		fmt.Fprintf(os.Stderr, "Missing name\n")
		valid = false
	}
	if !valid {
		return fmt.Errorf("object is not valid")
	}

	if stripRevision {
		object.Metadata.Id = ""
		object.Metadata.Revision = ""
	}

	var mode WriteMode
	if create && replace {
		mode = CreateOrReplace
	} else if create {
		mode = Create
	} else if replace {
		mode = Replace
	} else {
		return fmt.Errorf("Nothing to do")
	}

	client, err := GetClientFromEnv()
	if err != nil {
		return err
	}

	_, err = client.WriteObject(object, mode)
	return err
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
