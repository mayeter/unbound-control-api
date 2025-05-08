package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	rate       float64 // tokens per second
	bucketSize float64 // maximum bucket size
	clients    map[string]*clientLimiter
	mu         sync.RWMutex
}

type clientLimiter struct {
	tokens     float64
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, bucketSize float64) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		clients:    make(map[string]*clientLimiter),
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Get the first IP in the chain
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to remote address
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}

// refill adds tokens to the bucket based on elapsed time
func (rl *RateLimiter) refill(client *clientLimiter) {
	now := time.Now()
	elapsed := now.Sub(client.lastRefill).Seconds()
	client.tokens = min(rl.bucketSize, client.tokens+elapsed*rl.rate)
	client.lastRefill = now
}

// allow checks if a request can be processed
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Get or create client limiter
	client, exists := rl.clients[ip]
	if !exists {
		client = &clientLimiter{
			tokens:     rl.bucketSize,
			lastRefill: time.Now(),
		}
		rl.clients[ip] = client
	}

	rl.refill(client)
	if client.tokens >= 1 {
		client.tokens--
		return true
	}
	return false
}

// cleanup removes old client limiters
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, client := range rl.clients {
		// Remove clients that haven't been active for more than 1 hour
		if now.Sub(client.lastRefill) > time.Hour {
			delete(rl.clients, ip)
		}
	}
}

// RateLimit middleware limits the number of requests per second per client
func RateLimit(requestsPerSecond, burstSize float64) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerSecond, burstSize)

	// Start cleanup goroutine
	go func() {
		for {
			time.Sleep(time.Hour)
			limiter.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !limiter.allow(ip) {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
