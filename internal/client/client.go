// Package client connects CobaltDB commands to a TCP server.
package client

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/bcsamrudh/cobaltdb/internal/protocol"
)

// Client sends commands to one CobaltDB server.
type Client struct {
	address string
}

// New creates a client for address.
func New(address string) *Client {
	return &Client{address: address}
}

// Execute sends one command and returns its response.
func (c *Client) Execute(command string) (protocol.Response, error) {
	connection, err := net.Dial("tcp", c.address)
	if err != nil {
		return "", fmt.Errorf("connect to %s: %w", c.address, err)
	}
	defer connection.Close()

	return exchange(connection, command)
}

func exchange(connection net.Conn, command string) (protocol.Response, error) {
	if _, err := fmt.Fprintln(connection, command); err != nil {
		return "", fmt.Errorf("send command: %w", err)
	}

	response, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return protocol.Response(strings.TrimRight(response, "\r\n")), nil
}
