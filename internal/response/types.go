package response

// CommonResponse is the base response structure for all API responses
type CommonResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an API error response
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// StatusResponse represents the response from the status command
type StatusResponse struct {
	Version   string   `json:"version"`
	Verbosity int      `json:"verbosity"`
	Threads   int      `json:"threads"`
	Modules   []string `json:"modules"`
	Uptime    Uptime   `json:"uptime"`
	Options   Options  `json:"options"`
}

// Uptime represents server uptime information
type Uptime struct {
	Seconds   int    `json:"seconds"`
	Formatted string `json:"formatted"`
}

// Options represents server options
type Options struct {
	Control string `json:"control"`
}

// StatsResponse represents the response from the stats command
type StatsResponse struct {
	Queries     QueryStats       `json:"queries"`
	Cache       CacheStats       `json:"cache"`
	Recursion   RecursionStats   `json:"recursion"`
	RequestList RequestListStats `json:"request_list"`
	TCPUsage    float64          `json:"tcp_usage"`
}

// QueryStats represents query-related statistics
type QueryStats struct {
	Total         int `json:"total"`
	IPRateLimited int `json:"ip_ratelimited"`
}

// CacheStats represents cache-related statistics
type CacheStats struct {
	Hits     int `json:"hits"`
	Misses   int `json:"misses"`
	Prefetch int `json:"prefetch"`
	ZeroTTL  int `json:"zero_ttl"`
}

// RecursionStats represents recursion-related statistics
type RecursionStats struct {
	Replies int           `json:"replies"`
	Time    RecursionTime `json:"time"`
}

// RecursionTime represents recursion timing statistics
type RecursionTime struct {
	Average float64 `json:"average"`
	Median  float64 `json:"median"`
}

// RequestListStats represents request list statistics
type RequestListStats struct {
	Average     float64         `json:"average"`
	Max         int             `json:"max"`
	Overwritten int             `json:"overwritten"`
	Exceeded    int             `json:"exceeded"`
	Current     CurrentRequests `json:"current"`
}

// CurrentRequests represents current request statistics
type CurrentRequests struct {
	All  int `json:"all"`
	User int `json:"user"`
}
