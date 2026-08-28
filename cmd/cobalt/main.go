package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bcsamrudh/cobaltdb/internal/server"
	"github.com/bcsamrudh/cobaltdb/internal/store"
)

const usage = `Usage:
  cobalt server [-dev] [-addr address]

Commands:
  server    Start the CobaltDB server
  kv        Key-value commands (coming in a later task)
`

type serverOptions struct {
	address string
	dev     bool
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

		database := server.New(store.New())
		if options.dev {
			log.Printf("CobaltDB development server listening on %s", options.address)
		} else {
			log.Printf("CobaltDB server listening on %s", options.address)
		}
		return database.ListenAndServe(options.address)
	case "kv":
		return errors.New("kv commands are not implemented yet")
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func parseServerOptions(args []string) (serverOptions, error) {
	options := serverOptions{}
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&options.dev, "dev", false, "run the in-memory development server")
	flags.StringVar(&options.address, "addr", "127.0.0.1:6380", "TCP address to listen on")
	if err := flags.Parse(args); err != nil {
		return serverOptions{}, err
	}
	if flags.NArg() != 0 {
		return serverOptions{}, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	return options, nil
}
