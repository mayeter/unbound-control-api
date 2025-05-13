package zonefile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Record represents a DNS record in a zone file
type Record struct {
	Name     string
	TTL      int
	Class    string
	Type     string
	RData    string
	Comments string
}

// ZoneFile represents a DNS zone file
type ZoneFile struct {
	Name    string
	Records []Record
}

// LoadZoneFile loads a zone file from disk
func LoadZoneFile(filePath string) (*ZoneFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zone file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var records []Record
	var currentRecord *Record

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") {
			if currentRecord != nil && currentRecord.Comments == "" {
				currentRecord.Comments = strings.TrimPrefix(line, ";")
			}
			continue
		}

		// Parse record
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue // Skip invalid lines
		}

		record := Record{
			Name:  fields[0],
			Type:  fields[len(fields)-2],
			RData: fields[len(fields)-1],
		}

		// Parse TTL and Class if present
		if len(fields) > 3 {
			if ttl, err := parseTTL(fields[1]); err == nil {
				record.TTL = ttl
				record.Class = fields[2]
			} else {
				record.Class = fields[1]
			}
		}

		records = append(records, record)
		currentRecord = &records[len(records)-1]
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading zone file: %w", err)
	}

	return &ZoneFile{
		Name:    filepath.Base(filePath),
		Records: records,
	}, nil
}

// SaveZoneFile saves a zone file to disk
func SaveZoneFile(filePath string, zoneFile *ZoneFile) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create zone file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write SOA record first if it exists
	for _, record := range zoneFile.Records {
		if record.Type == "SOA" {
			writeRecord(writer, record)
			break
		}
	}

	// Write remaining records
	for _, record := range zoneFile.Records {
		if record.Type != "SOA" {
			writeRecord(writer, record)
		}
	}

	return writer.Flush()
}

// AddRecord adds a new record to the zone file
func (zf *ZoneFile) AddRecord(record Record) {
	zf.Records = append(zf.Records, record)
}

// RemoveRecord removes a record from the zone file
func (zf *ZoneFile) RemoveRecord(name, recordType string) {
	var newRecords []Record
	for _, record := range zf.Records {
		if record.Name != name || record.Type != recordType {
			newRecords = append(newRecords, record)
		}
	}
	zf.Records = newRecords
}

// UpdateRecord updates an existing record in the zone file
func (zf *ZoneFile) UpdateRecord(record Record) {
	for i, r := range zf.Records {
		if r.Name == record.Name && r.Type == record.Type {
			zf.Records[i] = record
			break
		}
	}
}

// GetRecord retrieves a specific record from the zone file
func (zf *ZoneFile) GetRecord(name, recordType string) *Record {
	for _, record := range zf.Records {
		if record.Name == name && record.Type == recordType {
			return &record
		}
	}
	return nil
}

// Helper functions
func writeRecord(writer *bufio.Writer, record Record) {
	if record.Comments != "" {
		fmt.Fprintf(writer, "; %s\n", record.Comments)
	}

	if record.TTL > 0 {
		fmt.Fprintf(writer, "%s\t%d\t%s\t%s\t%s\n",
			record.Name, record.TTL, record.Class, record.Type, record.RData)
	} else {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			record.Name, record.Class, record.Type, record.RData)
	}
}

func parseTTL(ttlStr string) (int, error) {
	var ttl int
	_, err := fmt.Sscanf(ttlStr, "%d", &ttl)
	return ttl, err
}
