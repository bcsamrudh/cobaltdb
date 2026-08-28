// Package server exposes CobaltDB over TCP.
package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/bcsamrudh/cobaltdb/internal/store"
)

// Server handles TCP clients using a shared Store.
type Server struct {
	store *store.Store
}

// New creates a server backed by database.
func New(database *store.Store) *Server {
	return &Server{store: database}
}

// ListenAndServe listens on address and serves clients until it encounters an
// error.
func (s *Server) ListenAndServe(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	defer listener.Close()
	return s.Serve(listener)
}

// Serve accepts connections from listener. Each client is handled in its own
// goroutine.
func (s *Server) Serve(listener net.Listener) error {
	for {
		connection, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("accept connection: %w", err)
		}
		go s.handleConnection(connection)
	}
}

func (s *Server) handleConnection(connection net.Conn) {
	defer connection.Close()

	scanner := bufio.NewScanner(connection)
	writer := bufio.NewWriter(connection)
	for scanner.Scan() {
		if _, err := fmt.Fprintln(writer, s.execute(scanner.Text())); err != nil {
			return
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}

// execute provides the minimal command handling needed for Day 3. The formal
// CobaltDB protocol will replace this parser in the next task.
func (s *Server) execute(line string) string {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "ERROR empty command"
	}

	switch strings.ToUpper(parts[0]) {
	case "SET":
		if len(parts) != 3 {
			return "ERROR usage: SET <key> <value>"
		}
		s.store.Set(parts[1], parts[2])
		return "OK"
	case "GET":
		if len(parts) != 2 {
			return "ERROR usage: GET <key>"
		}
		value, exists := s.store.Get(parts[1])
		if !exists {
			return "(nil)"
		}
		return value
	case "DELETE":
		if len(parts) != 2 {
			return "ERROR usage: DELETE <key>"
		}
		if !s.store.Delete(parts[1]) {
			return "(nil)"
		}
		return "OK"
	case "EXISTS":
		if len(parts) != 2 {
			return "ERROR usage: EXISTS <key>"
		}
		if s.store.Exists(parts[1]) {
			return "true"
		}
		return "false"
	default:
		return "ERROR unknown command"
	}
}
