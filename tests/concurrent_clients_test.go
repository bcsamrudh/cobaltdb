package tests

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bcsamrudh/cobaltdb/internal/server"
	"github.com/bcsamrudh/cobaltdb/internal/store"
)

func TestConcurrentTCPClients(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.New(store.New()).Serve(listener)
	}()

	const (
		clients    = 100
		iterations = 100
	)

	start := make(chan struct{})
	errors := make(chan error, clients)
	var clientsDone sync.WaitGroup
	clientsDone.Add(clients)
	for clientID := 0; clientID < clients; clientID++ {
		go func(clientID int) {
			defer clientsDone.Done()
			<-start
			if err := runWorkload(listener.Addr().String(), clientID, iterations); err != nil {
				errors <- err
			}
		}(clientID)
	}

	close(start)
	clientsDone.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}

	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-serveDone; err != nil {
		t.Fatalf("server stopped with an error: %v", err)
	}
}

func runWorkload(address string, clientID, iterations int) error {
	connection, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("client %d connect: %w", clientID, err)
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return fmt.Errorf("client %d set deadline: %w", clientID, err)
	}

	reader := bufio.NewReader(connection)
	key := fmt.Sprintf("client-%d", clientID)
	for iteration := 0; iteration < iterations; iteration++ {
		value := fmt.Sprintf("value-%d", iteration)
		if err := exchange(connection, reader, "SET "+key+" "+value, "OK"); err != nil {
			return fmt.Errorf("client %d iteration %d: %w", clientID, iteration, err)
		}
		if err := exchange(connection, reader, "GET "+key, "VALUE "+value); err != nil {
			return fmt.Errorf("client %d iteration %d: %w", clientID, iteration, err)
		}
	}

	if err := exchange(connection, reader, "EXISTS "+key, "EXISTS"); err != nil {
		return fmt.Errorf("client %d: %w", clientID, err)
	}
	if err := exchange(connection, reader, "DELETE "+key, "DELETED"); err != nil {
		return fmt.Errorf("client %d: %w", clientID, err)
	}
	if err := exchange(connection, reader, "GET "+key, "NOT_FOUND"); err != nil {
		return fmt.Errorf("client %d: %w", clientID, err)
	}
	return nil
}

func exchange(connection net.Conn, reader *bufio.Reader, request, want string) error {
	if _, err := fmt.Fprintln(connection, request); err != nil {
		return fmt.Errorf("send %q: %w", request, err)
	}
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read response for %q: %w", request, err)
	}
	if got := strings.TrimSpace(response); got != want {
		return fmt.Errorf("%q returned %q, want %q", request, got, want)
	}
	return nil
}
