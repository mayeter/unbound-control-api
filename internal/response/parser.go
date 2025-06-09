package response

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseStatusResponse parses the raw status command response into a StatusResponse
func ParseStatusResponse(raw string) (*StatusResponse, error) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	status := &StatusResponse{}

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "version":
			status.Version = value
		case "verbosity":
			if v, err := strconv.Atoi(value); err == nil {
				status.Verbosity = v
			}
		case "threads":
			if v, err := strconv.Atoi(value); err == nil {
				status.Threads = v
			}
		case "modules":
			// Remove brackets and split by space
			modules := strings.Trim(value, "[]")
			status.Modules = strings.Fields(modules)
		case "uptime":
			// Parse uptime in seconds
			if v, err := strconv.Atoi(strings.Split(value, " ")[0]); err == nil {
				status.Uptime.Seconds = v
				status.Uptime.Formatted = formatUptime(v)
			}
		case "options":
			// Parse options like "control(open)"
			if strings.Contains(value, "control") {
				status.Options.Control = "open"
			}
		}
	}

	return status, nil
}

// ParseStatsResponse parses the raw stats command response into a StatsResponse
func ParseStatsResponse(raw string) (*StatsResponse, error) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	stats := &StatsResponse{}

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse the value
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}

		// Map the key to the appropriate field
		switch {
		case strings.HasPrefix(key, "total.num.queries"):
			stats.Queries.Total = int(val)
		case strings.HasPrefix(key, "total.num.queries_ip_ratelimited"):
			stats.Queries.IPRateLimited = int(val)
		case strings.HasPrefix(key, "total.num.cachehits"):
			stats.Cache.Hits = int(val)
		case strings.HasPrefix(key, "total.num.cachemiss"):
			stats.Cache.Misses = int(val)
		case strings.HasPrefix(key, "total.num.prefetch"):
			stats.Cache.Prefetch = int(val)
		case strings.HasPrefix(key, "total.num.zero_ttl"):
			stats.Cache.ZeroTTL = int(val)
		case strings.HasPrefix(key, "total.num.recursivereplies"):
			stats.Recursion.Replies = int(val)
		case strings.HasPrefix(key, "total.requestlist.avg"):
			stats.RequestList.Average = val
		case strings.HasPrefix(key, "total.requestlist.max"):
			stats.RequestList.Max = int(val)
		case strings.HasPrefix(key, "total.requestlist.overwritten"):
			stats.RequestList.Overwritten = int(val)
		case strings.HasPrefix(key, "total.requestlist.exceeded"):
			stats.RequestList.Exceeded = int(val)
		case strings.HasPrefix(key, "total.requestlist.current.all"):
			stats.RequestList.Current.All = int(val)
		case strings.HasPrefix(key, "total.requestlist.current.user"):
			stats.RequestList.Current.User = int(val)
		case strings.HasPrefix(key, "total.recursion.time.avg"):
			stats.Recursion.Time.Average = val
		case strings.HasPrefix(key, "total.recursion.time.median"):
			stats.Recursion.Time.Median = val
		case strings.HasPrefix(key, "total.tcpusage"):
			stats.TCPUsage = val
		}
	}

	return stats, nil
}

// formatUptime converts seconds into a human-readable duration string
func formatUptime(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
