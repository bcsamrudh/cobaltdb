package client

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/bcsamrudh/cobaltdb/internal/protocol"
)

func TestExchange(t *testing.T) {
	serverConnection, clientConnection := net.Pipe()
	defer clientConnection.Close()

	serverDone := make(chan error, 1)
	go func() {
		defer serverConnection.Close()
		command, err := bufio.NewReader(serverConnection).ReadString('\n')
		if err != nil {
			serverDone <- err
			return
		}
		if command != "GET name\n" {
			serverDone <- fmt.Errorf("command = %q, want %q", command, "GET name\n")
			return
		}
		_, err = fmt.Fprintln(serverConnection, "VALUE samrudh")
		serverDone <- err
	}()

	response, err := exchange(clientConnection, "GET name")
	if err != nil {
		t.Fatal(err)
	}
	if response != protocol.Value("samrudh") {
		t.Fatalf("response = %q, want %q", response, protocol.Value("samrudh"))
	}
	if err := <-serverDone; err != nil {
		t.Fatal(err)
	}
}
