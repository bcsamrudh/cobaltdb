// Package protocol implements CobaltDB's line-oriented text protocol.
package protocol

import (
	"errors"
	"strings"
)

// Operation identifies a command supported by CobaltDB.
type Operation string

const (
	Set    Operation = "SET"
	Get    Operation = "GET"
	Delete Operation = "DELETE"
	Exists Operation = "EXISTS"
)

// Command is a parsed client request.
type Command struct {
	Operation Operation
	Key       string
	Value     string
}

// Store describes the storage operations required by the protocol.
type Store interface {
	Set(key, value string) error
	Get(key string) (string, bool)
	Delete(key string) (bool, error)
	Exists(key string) bool
}

// Response is one newline-delimited response sent to a client.
type Response string

const (
	OK        Response = "OK"
	NotFound  Response = "NOT_FOUND"
	Deleted   Response = "DELETED"
	ExistsYes Response = "EXISTS"
	ExistsNo  Response = "NOT_EXISTS"
)

// Parse converts one line of client input into a Command. Command names are
// case-insensitive, keys cannot contain spaces, and SET values may contain
// spaces.
func Parse(line string) (Command, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Command{}, errors.New("empty command")
	}

	operation := Operation(strings.ToUpper(fields[0]))
	switch operation {
	case Set:
		if len(fields) < 3 {
			return Command{}, errors.New("usage: SET <key> <value>")
		}
		return Command{Operation: Set, Key: fields[1], Value: strings.Join(fields[2:], " ")}, nil
	case Get, Delete, Exists:
		if len(fields) != 2 {
			return Command{}, errors.New("usage: " + string(operation) + " <key>")
		}
		return Command{Operation: operation, Key: fields[1]}, nil
	default:
		return Command{}, errors.New("unknown command: " + fields[0])
	}
}

// Execute parses a request, applies it to database, and returns its wire
// response.
func Execute(database Store, line string) Response {
	command, err := Parse(line)
	if err != nil {
		return Error(err)
	}

	switch command.Operation {
	case Set:
		if err := database.Set(command.Key, command.Value); err != nil {
			return Error(err)
		}
		return OK
	case Get:
		value, found := database.Get(command.Key)
		if !found {
			return NotFound
		}
		return Value(value)
	case Delete:
		deleted, err := database.Delete(command.Key)
		if err != nil {
			return Error(err)
		}
		if !deleted {
			return NotFound
		}
		return Deleted
	case Exists:
		if database.Exists(command.Key) {
			return ExistsYes
		}
		return ExistsNo
	default:
		return Error(errors.New("unsupported command"))
	}
}

// Value creates a successful GET response.
func Value(value string) Response {
	return Response("VALUE " + value)
}

// Error creates an error response.
func Error(err error) Response {
	return Response("ERROR " + err.Error())
}
