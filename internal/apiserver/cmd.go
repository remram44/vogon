package apiserver

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"

	"github.com/remram44/vogon/internal/commands"
	"github.com/remram44/vogon/internal/database"
)

func transmute(name string, source interface{}, dest interface{}) error {
	metadata := mapstructure.Metadata{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:   &dest,
		TagName:  "yaml",
		Metadata: &metadata,
	})
	if err != nil {
		return err
	}
	err = decoder.Decode(source)
	if err != nil {
		return err
	}
	if len(metadata.Unused) > 0 {
		return fmt.Errorf("unexpected keys in %v: %v", name, strings.Join(metadata.Unused, ", "))
	}
	return nil
}

type Config struct {
	ListenAddr string                `yaml:"listen_addr"`
	ListenPort int                   `yaml:"listen_port"`
	Database   DatabaseConfigWrapper `yaml:"database"`
}

type DatabaseConfig interface {
	Connect() (database.Database, error)
}

type DatabaseConfigWrapper struct {
	DatabaseConfig
}

type InMemoryDatabaseConfig struct {
}

func (*InMemoryDatabaseConfig) Connect() (database.Database, error) {
	log.Print("open InMemoryDatabase")
	return database.NewInMemoryDatabase(), nil
}

type EtcdDatabaseConfig struct {
	Hostname   string `yaml:"hostname"`
	CaCert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"client_cert"`
	ClientKey  string `yaml:"client_key"`
}

func (db *EtcdDatabaseConfig) Connect() (database.Database, error) {
	log.Print("open EtcdDatabase")
	return nil, fmt.Errorf("Not implemented")
}

func (db *DatabaseConfigWrapper) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	typeValue, ok := raw["type"]
	if !ok {
		return fmt.Errorf("Missing 'type' in DatabaseConfigWrapper")
	}
	typeString, ok := typeValue.(string)
	if !ok {
		return fmt.Errorf("'type' is not a string")
	}
	delete(raw, "type")

	switch typeString {
	case "in_memory":
		var finalValue InMemoryDatabaseConfig
		if err := transmute("InMemoryDatabaseConfig", raw, &finalValue); err != nil {
			return err
		}
		db.DatabaseConfig = &finalValue
	case "etcd":
		var finalValue EtcdDatabaseConfig
		if err := transmute("EtcdDatabaseConfig", raw, &finalValue); err != nil {
			return err
		}
		db.DatabaseConfig = &finalValue
	default:
		return fmt.Errorf("Unknown database type %v", typeString)
	}
	return nil
}

func PrintUsage(w io.Writer) {
	fmt.Fprintf(
		w,
		""+
			"  apiserver <config>\n"+
			"    Run the API server\n",
	)
}

func Run(args []string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Missing config file\n")
		os.Exit(2)
	}

	f, err := os.Open(args[1])
	if err != nil {
		log.Fatalf("opening config file: %v", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("parsing config file: %v", err)
	}

	err = runServer(config)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	commands.Register("apiserver", &commands.Command{
		PrintUsage: PrintUsage,
		Run:        Run,
	})
}
