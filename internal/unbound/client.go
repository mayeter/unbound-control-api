package unbound

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewClient(host string, port int, certFile, keyFile string) (*Client, error) {
	config := &tls.Config{
		InsecureSkipVerify: true, // In production, set this to false and provide proper certificates
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to unbound control: %w", err)
	}

	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *Client) SendCommand(cmd string) (string, error) {
	// Send command
	if _, err := fmt.Fprintf(c.conn, "%s\n", cmd); err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	response, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "error") {
		return "", fmt.Errorf("unbound control error: %s", response)
	}

	return response, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Common commands
func (c *Client) Status() (string, error) {
	return c.SendCommand("status")
}

func (c *Client) Reload() (string, error) {
	return c.SendCommand("reload")
}

func (c *Client) Flush() (string, error) {
	return c.SendCommand("flush")
}

func (c *Client) Stats() (string, error) {
	return c.SendCommand("stats")
}

func (c *Client) Info() (string, error) {
	return c.SendCommand("info")
}
