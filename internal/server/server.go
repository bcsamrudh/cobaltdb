// Package server exposes CobaltDB over TCP.
package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"

	"github.com/bcsamrudh/cobaltdb/internal/protocol"
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
		response := protocol.Execute(s.store, scanner.Text())
		if _, err := fmt.Fprintln(writer, response); err != nil {
			return
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}
