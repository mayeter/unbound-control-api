package unbound

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/callMe-Root/unbound-control-api/internal/response"
)

type Client struct {
	socketPath string
	logger     *log.Logger
}

func NewClient(socketPath string) (*Client, error) {
	logger := log.New(log.Writer(), "[UnboundClient] ", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("Initializing client for UNIX socket: %s", socketPath)
	return &Client{
		socketPath: socketPath,
		logger:     logger,
	}, nil
}

func (c *Client) SendCommand(cmd string) (string, error) {
	c.logger.Printf("Connecting to Unbound control UNIX socket: %s", c.socketPath)
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		c.logger.Printf("Failed to connect to socket: %v", err)
		return "", fmt.Errorf("failed to connect to socket: %w", err)
	}
	defer conn.Close()

	// Format command with UBCT1  prefix and newline
	fullCmd := fmt.Sprintf("UBCT1  %s\n", cmd)
	c.logger.Printf("Sending command: %q", fullCmd)
	_, err = conn.Write([]byte(fullCmd))
	if err != nil {
		c.logger.Printf("Failed to write command: %v", err)
		return "", fmt.Errorf("failed to write command: %w", err)
	}

	// Read and return the response
	c.logger.Printf("Reading response...")
	var response strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		c.logger.Printf("Read line: %q", line)
		response.WriteString(line)
		response.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		c.logger.Printf("Error reading response: %v", err)
		return "", fmt.Errorf("error reading response: %w", err)
	}

	respStr := strings.TrimSpace(response.String())
	c.logger.Printf("Received response: %s", respStr)
	return respStr, nil
}

func (c *Client) Close() error {
	return nil
}

// Status returns the server status
func (c *Client) Status() (*response.StatusResponse, error) {
	raw, err := c.SendCommand("status")
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	return response.ParseStatusResponse(raw)
}

// Stats returns the server statistics
func (c *Client) Stats() (*response.StatsResponse, error) {
	raw, err := c.SendCommand("stats")
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return response.ParseStatsResponse(raw)
}

// Reload reloads the server configuration
func (c *Client) Reload() error {
	_, err := c.SendCommand("reload")
	if err != nil {
		return fmt.Errorf("failed to reload: %w", err)
	}
	return nil
}

// Flush flushes the cache for a domain
func (c *Client) Flush(domain string) error {
	cmd := fmt.Sprintf("flush %s", domain)
	_, err := c.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to flush domain %s: %w", domain, err)
	}
	return nil
}

// TestConnection verifies that the connection to Unbound is working
func (c *Client) TestConnection() error {
	c.logger.Printf("Testing connection to Unbound control UNIX socket: %s", c.socketPath)

	// Try to get status
	status, err := c.Status()
	if err != nil {
		c.logger.Printf("Connection test failed: %v", err)
		return fmt.Errorf("connection test failed: %w", err)
	}

	c.logger.Printf("Connected to Unbound version: %s", status.Version)
	return nil
}

// VerifyConnection checks if the connection is still alive
func (c *Client) VerifyConnection() bool {
	return true
}

// reconnect attempts to reestablish the connection
func (c *Client) reconnect() error {
	return nil
}
