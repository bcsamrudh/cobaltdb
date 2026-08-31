package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/bcsamrudh/cobaltdb/internal/store"
)

func TestConnectionHandlesMultipleCommands(t *testing.T) {
	s := New(store.New())
	serverConnection, clientConnection := net.Pipe()
	go s.handleConnection(serverConnection)
	defer clientConnection.Close()

	reader := bufio.NewReader(clientConnection)
	requests := []struct {
		command string
		want    string
	}{
		{"SET language go", "OK"},
		{"GET language", "VALUE go"},
		{"DELETE language", "DELETED"},
		{"GET language", "NOT_FOUND"},
	}

	for _, request := range requests {
		if _, err := fmt.Fprintln(clientConnection, request.command); err != nil {
			t.Fatal(err)
		}
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.TrimSpace(response); got != request.want {
			t.Fatalf("%q returned %q, want %q", request.command, got, request.want)
		}
	}
}

func TestConcurrentConnectionsShareStore(t *testing.T) {
	database := store.New()
	s := New(database)

	const clients = 20
	var wg sync.WaitGroup
	wg.Add(clients)
	for client := 0; client < clients; client++ {
		go func(client int) {
			defer wg.Done()
			serverConnection, clientConnection := net.Pipe()
			go s.handleConnection(serverConnection)
			defer clientConnection.Close()

			key := fmt.Sprintf("client-%d", client)
			if _, err := fmt.Fprintf(clientConnection, "SET %s %d\n", key, client); err != nil {
				t.Errorf("client %d write: %v", client, err)
				return
			}
			response, err := bufio.NewReader(clientConnection).ReadString('\n')
			if err != nil {
				t.Errorf("client %d read: %v", client, err)
				return
			}
			if strings.TrimSpace(response) != "OK" {
				t.Errorf("client %d response = %q, want OK", client, response)
			}
		}(client)
	}
	wg.Wait()

	for client := 0; client < clients; client++ {
		key := fmt.Sprintf("client-%d", client)
		want := fmt.Sprint(client)
		if got, exists := database.Get(key); !exists || got != want {
			t.Errorf("Get(%s) = %q, %v; want %q, true", key, got, exists, want)
		}
	}
}
