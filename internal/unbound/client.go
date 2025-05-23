package unbound

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/callMe-Root/unbound-control-api/internal/zonefile"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	addr   string
	config *tls.Config
	logger *log.Logger
}

// Zone represents a DNS zone configuration
type Zone struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"` // primary, secondary, stub, forward
	File     string   `json:"file,omitempty"`
	Masters  []string `json:"masters,omitempty"`
	Forwards []string `json:"forwards,omitempty"`
}

// ZoneResponse represents the response from zone-related commands
type ZoneResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Zones   []Zone `json:"zones,omitempty"`
}

func NewClient(host string, port int, certFile, keyFile string) (*Client, error) {
	// Create logger
	logger := log.New(log.Writer(), "[UnboundClient] ", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("Initializing client for %s:%d", host, port)

	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.Printf("Failed to load certificates: %v", err)
		return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
	}
	logger.Printf("Successfully loaded certificates from %s and %s", certFile, keyFile)

	// Create TLS config with more lenient settings for containerized Unbound
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true, // Skip certificate verification for containerized setup
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		VerifyConnection: func(cs tls.ConnectionState) error {
			logger.Printf("TLS Connection State:")
			logger.Printf("  Version: %d", cs.Version)
			logger.Printf("  HandshakeComplete: %v", cs.HandshakeComplete)
			logger.Printf("  DidResume: %v", cs.DidResume)
			logger.Printf("  CipherSuite: %d", cs.CipherSuite)
			logger.Printf("  NegotiatedProtocol: %s", cs.NegotiatedProtocol)
			logger.Printf("  ServerName: %s", cs.ServerName)
			logger.Printf("  PeerCertificates: %d", len(cs.PeerCertificates))
			return nil
		},
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	logger.Printf("Attempting to connect to %s", addr)

	// Connect to Unbound control interface with a dialer that includes keepalive
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// First try a plain TCP connection to verify basic connectivity
	tcpConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		logger.Printf("TCP connection failed: %v", err)
		return nil, fmt.Errorf("failed to establish TCP connection: %w", err)
	}
	logger.Printf("TCP connection successful, upgrading to TLS")
	tcpConn.Close()

	// Now try the TLS connection
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, config)
	if err != nil {
		logger.Printf("TLS connection failed: %v", err)
		return nil, fmt.Errorf("failed to establish TLS connection: %w", err)
	}

	// Verify the handshake completed
	if !conn.ConnectionState().HandshakeComplete {
		conn.Close()
		logger.Printf("TLS handshake did not complete")
		return nil, fmt.Errorf("TLS handshake did not complete")
	}

	logger.Printf("Successfully connected to %s", addr)

	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		addr:   addr,
		config: config,
		logger: logger,
	}

	// Test the connection
	if err := client.TestConnection(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	return client, nil
}

func (c *Client) SendCommand(cmd string) (string, error) {
	c.logger.Printf("Sending command: %s", cmd)

	// Verify connection before sending
	if !c.VerifyConnection() {
		c.logger.Printf("Connection appears to be dead, attempting to reconnect")
		if err := c.reconnect(); err != nil {
			c.logger.Printf("Reconnection failed: %v", err)
			return "", fmt.Errorf("failed to reconnect: %w", err)
		}
	}

	// Send command with newline
	if _, err := fmt.Fprintf(c.conn, "%s\n", cmd); err != nil {
		c.logger.Printf("Failed to send command: %v", err)
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Ensure the command is sent immediately
	if err := c.conn.(*tls.Conn).SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		c.logger.Printf("Failed to set write deadline: %v", err)
	}

	c.logger.Printf("Command sent successfully")

	// Read response with timeout
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// For Unbound 1.22, we need to read until we get a blank line or EOF
	var response strings.Builder
	var emptyLines int
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				if response.Len() > 0 {
					// EOF after receiving some data is okay
					break
				}
				// Check if the connection is still alive
				if !c.VerifyConnection() {
					c.logger.Printf("Connection died after sending command")
					return "", fmt.Errorf("connection died after sending command")
				}
				c.logger.Printf("Received EOF without any data, connection still alive")
				// Try to read one more time with a shorter timeout
				c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				line, err = c.reader.ReadString('\n')
				if err != nil {
					c.logger.Printf("Still no data after retry: %v", err)
					return "", fmt.Errorf("received EOF without any data")
				}
				// If we got data, continue processing
				line = strings.TrimSpace(line)
				if line != "" {
					response.WriteString(line)
					response.WriteString("\n")
					continue
				}
				return "", fmt.Errorf("received EOF without any data")
			}
			c.logger.Printf("Failed to read response: %v", err)
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		// Trim the line and check if it's empty
		line = strings.TrimSpace(line)
		if line == "" {
			emptyLines++
			if emptyLines >= 2 {
				// Two consecutive empty lines indicate end of response
				break
			}
			continue
		}
		emptyLines = 0

		response.WriteString(line)
		response.WriteString("\n")
	}

	c.conn.SetReadDeadline(time.Time{}) // Reset deadline

	responseStr := strings.TrimSpace(response.String())
	if responseStr == "" {
		c.logger.Printf("Received empty response")
		return "", fmt.Errorf("received empty response")
	}

	c.logger.Printf("Received response: %s", responseStr)

	if strings.HasPrefix(responseStr, "error") {
		c.logger.Printf("Unbound control error: %s", responseStr)
		return "", fmt.Errorf("unbound control error: %s", responseStr)
	}

	return responseStr, nil
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

// Zone management commands
func (c *Client) ListZones() ([]Zone, error) {
	response, err := c.SendCommand("list_zones")
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	// Parse the response into Zone structs
	var zones []Zone
	if err := json.Unmarshal([]byte(response), &zones); err != nil {
		return nil, fmt.Errorf("failed to parse zones: %w", err)
	}

	return zones, nil
}

func (c *Client) AddZone(zone Zone) error {
	// Convert zone to JSON
	zoneJSON, err := json.Marshal(zone)
	if err != nil {
		return fmt.Errorf("failed to marshal zone: %w", err)
	}

	// Send add_zone command with zone data
	cmd := fmt.Sprintf("add_zone %s", string(zoneJSON))
	_, err = c.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to add zone: %w", err)
	}

	return nil
}

func (c *Client) RemoveZone(zoneName string) error {
	cmd := fmt.Sprintf("remove_zone %s", zoneName)
	_, err := c.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove zone: %w", err)
	}

	return nil
}

func (c *Client) UpdateZone(zone Zone) error {
	// Convert zone to JSON
	zoneJSON, err := json.Marshal(zone)
	if err != nil {
		return fmt.Errorf("failed to marshal zone: %w", err)
	}

	// Send update_zone command with zone data
	cmd := fmt.Sprintf("update_zone %s", string(zoneJSON))
	_, err = c.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to update zone: %w", err)
	}

	return nil
}

func (c *Client) GetZone(zoneName string) (*Zone, error) {
	cmd := fmt.Sprintf("get_zone %s", zoneName)
	response, err := c.SendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}

	var zone Zone
	if err := json.Unmarshal([]byte(response), &zone); err != nil {
		return nil, fmt.Errorf("failed to parse zone: %w", err)
	}

	return &zone, nil
}

// Zone file management commands
func (c *Client) GetZoneFile(zoneName string) (*zonefile.ZoneFile, error) {
	// Get zone configuration to find the file path
	zone, err := c.GetZone(zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone configuration: %w", err)
	}

	if zone.File == "" {
		return nil, fmt.Errorf("zone %s does not have a file path configured", zoneName)
	}

	// Load zone file from disk
	return zonefile.LoadZoneFile(zone.File)
}

func (c *Client) UpdateZoneFile(zoneName string, zoneFile *zonefile.ZoneFile) error {
	// Get zone configuration to find the file path
	zone, err := c.GetZone(zoneName)
	if err != nil {
		return fmt.Errorf("failed to get zone configuration: %w", err)
	}

	if zone.File == "" {
		return fmt.Errorf("zone %s does not have a file path configured", zoneName)
	}

	// Save zone file to disk
	if err := zonefile.SaveZoneFile(zone.File, zoneFile); err != nil {
		return fmt.Errorf("failed to save zone file: %w", err)
	}

	// Reload Unbound to apply changes
	_, err = c.Reload()
	if err != nil {
		return fmt.Errorf("failed to reload Unbound after zone file update: %w", err)
	}

	return nil
}

func (c *Client) AddZoneRecord(zoneName string, record zonefile.Record) error {
	// Get current zone file
	zoneFile, err := c.GetZoneFile(zoneName)
	if err != nil {
		return fmt.Errorf("failed to get zone file: %w", err)
	}

	// Add record
	zoneFile.AddRecord(record)

	// Save updated zone file
	return c.UpdateZoneFile(zoneName, zoneFile)
}

func (c *Client) RemoveZoneRecord(zoneName string, recordName string, recordType string) error {
	// Get current zone file
	zoneFile, err := c.GetZoneFile(zoneName)
	if err != nil {
		return fmt.Errorf("failed to get zone file: %w", err)
	}

	// Remove record
	zoneFile.RemoveRecord(recordName, recordType)

	// Save updated zone file
	return c.UpdateZoneFile(zoneName, zoneFile)
}

func (c *Client) UpdateZoneRecord(zoneName string, record zonefile.Record) error {
	// Get current zone file
	zoneFile, err := c.GetZoneFile(zoneName)
	if err != nil {
		return fmt.Errorf("failed to get zone file: %w", err)
	}

	// Update record
	zoneFile.UpdateRecord(record)

	// Save updated zone file
	return c.UpdateZoneFile(zoneName, zoneFile)
}

func (c *Client) GetZoneRecord(zoneName string, recordName string, recordType string) (*zonefile.Record, error) {
	// Get current zone file
	zoneFile, err := c.GetZoneFile(zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone file: %w", err)
	}

	// Get record
	record := zoneFile.GetRecord(recordName, recordType)
	if record == nil {
		return nil, fmt.Errorf("record not found")
	}

	return record, nil
}

// TestConnection verifies that the connection to Unbound is working
func (c *Client) TestConnection() error {
	c.logger.Printf("Testing connection to Unbound control interface")

	// Try to get status
	response, err := c.SendCommand("status")
	if err != nil {
		c.logger.Printf("Connection test failed: %v", err)
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Parse version from response
	version := ""
	for _, line := range strings.Split(response, "\n") {
		if strings.HasPrefix(line, "version:") {
			version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
			break
		}
	}

	if version == "" {
		c.logger.Printf("Could not determine Unbound version from response: %s", response)
		return fmt.Errorf("could not determine Unbound version")
	}

	c.logger.Printf("Connected to Unbound version: %s", version)

	// Verify response format for Unbound 1.22
	if !strings.Contains(response, "version:") || !strings.Contains(response, "threads:") {
		c.logger.Printf("Unexpected response format: %s", response)
		return fmt.Errorf("unexpected response format: %s", response)
	}

	c.logger.Printf("Connection test successful: %s", response)
	return nil
}

// VerifyConnection checks if the connection is still alive
func (c *Client) VerifyConnection() bool {
	if c.conn == nil {
		return false
	}

	// Try to get a one-byte read deadline
	c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	one := []byte{0}
	_, err := c.conn.Read(one)
	c.conn.SetReadDeadline(time.Time{}) // Reset deadline

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// Timeout is expected, connection is still alive
			return true
		}
		// Other errors indicate connection is dead
		return false
	}

	return true
}

// reconnect attempts to reestablish the connection
func (c *Client) reconnect() error {
	c.logger.Printf("Attempting to reconnect to %s", c.addr)

	// Close existing connection
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Printf("Error closing existing connection: %v", err)
		}
	}

	// Create new connection with keepalive
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", c.addr, c.config)
	if err != nil {
		c.logger.Printf("Reconnection failed: %v", err)
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.logger.Printf("Successfully reconnected to %s", c.addr)
	return nil
}
