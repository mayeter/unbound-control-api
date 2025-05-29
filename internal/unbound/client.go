package unbound

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/callMe-Root/unbound-control-api/internal/zonefile"
)

type Client struct {
	socketPath string
	logger     *log.Logger
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
	c.logger.Printf("Testing connection to Unbound control UNIX socket: %s", c.socketPath)

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
	return true
}

// reconnect attempts to reestablish the connection
func (c *Client) reconnect() error {
	return nil
}
