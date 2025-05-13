package unbound

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/callMe-Root/unbound-control-api/internal/zonefile"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
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
