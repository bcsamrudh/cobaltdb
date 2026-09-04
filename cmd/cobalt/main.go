package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bcsamrudh/cobaltdb/internal/client"
	"github.com/bcsamrudh/cobaltdb/internal/protocol"
	"github.com/bcsamrudh/cobaltdb/internal/server"
	"github.com/bcsamrudh/cobaltdb/internal/storage"
	"github.com/bcsamrudh/cobaltdb/internal/store"
)

const usage = `Usage:
  cobalt server [-dev] [-addr address] [-data-dir path]
  cobalt kv [-addr address] put <key> <value>
  cobalt kv [-addr address] get <key>
  cobalt kv [-addr address] delete <key>
  cobalt kv [-addr address] exists <key>

Commands:
  server    Start the CobaltDB server
  kv        Read and write key-value data
`

const defaultServerAddress = "127.0.0.1:6380"

type serverOptions struct {
	address       string
	dataDirectory string
	dev           bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Printf("error: %v", err)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("a command is required")
	}

	switch args[0] {
	case "server":
		options, err := parseServerOptions(args[1:])
		if err != nil {
			return err
		}

		var database protocol.Store
		if options.dev {
			database = store.New()
			log.Printf("CobaltDB development server listening on %s", options.address)
		} else {
			persistent, err := storage.Open(options.dataDirectory)
			if err != nil {
				return err
			}
			defer persistent.Close()
			database = persistent
			log.Printf("CobaltDB server listening on %s (data: %s)", options.address, options.dataDirectory)
		}
		return server.New(database).ListenAndServe(options.address)
	case "kv":
		return runKV(args[1:], os.Stdout)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runKV(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("kv", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	address := flags.String("addr", clientAddress(), "CobaltDB server address")
	if err := flags.Parse(args); err != nil {
		return err
	}

	request, err := buildKVRequest(flags.Args())
	if err != nil {
		return err
	}
	response, err := client.New(*address).Execute(request)
	if err != nil {
		return err
	}
	display, err := displayResponse(response)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(output, display)
	return err
}

func buildKVRequest(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("a kv operation is required")
	}

	switch strings.ToLower(args[0]) {
	case "put":
		if len(args) < 3 {
			return "", errors.New("usage: cobalt kv put <key> <value>")
		}
		return fmt.Sprintf("%s %s %s", protocol.Set, args[1], strings.Join(args[2:], " ")), nil
	case "get":
		return singleKeyRequest(protocol.Get, args)
	case "delete":
		return singleKeyRequest(protocol.Delete, args)
	case "exists":
		return singleKeyRequest(protocol.Exists, args)
	default:
		return "", fmt.Errorf("unknown kv operation %q", args[0])
	}
}

func singleKeyRequest(operation protocol.Operation, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("usage: cobalt kv %s <key>", strings.ToLower(string(operation)))
	}
	return fmt.Sprintf("%s %s", operation, args[1]), nil
}

func displayResponse(response protocol.Response) (string, error) {
	switch {
	case response == protocol.OK:
		return "OK", nil
	case response == protocol.Deleted:
		return "OK", nil
	case response == protocol.NotFound:
		return "(nil)", nil
	case response == protocol.ExistsYes:
		return "true", nil
	case response == protocol.ExistsNo:
		return "false", nil
	case strings.HasPrefix(string(response), "VALUE "):
		return strings.TrimPrefix(string(response), "VALUE "), nil
	case strings.HasPrefix(string(response), "ERROR "):
		return "", errors.New(strings.TrimPrefix(string(response), "ERROR "))
	default:
		return "", fmt.Errorf("invalid server response %q", response)
	}
}

func parseServerOptions(args []string) (serverOptions, error) {
	options := serverOptions{}
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&options.dev, "dev", false, "run the in-memory development server")
	flags.StringVar(&options.address, "addr", defaultServerAddress, "TCP address to listen on")
	flags.StringVar(&options.dataDirectory, "data-dir", "./data", "directory for persistent data")
	if err := flags.Parse(args); err != nil {
		return serverOptions{}, err
	}
	if flags.NArg() != 0 {
		return serverOptions{}, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	return options, nil
}

func clientAddress() string {
	if address := os.Getenv("COBALT_ADDR"); address != "" {
		return address
	}
	return defaultServerAddress
}
