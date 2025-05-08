package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/callMe-Root/unbound-control-api/pkg/logger"
	"github.com/rs/zerolog"
)

// responseWriter is a custom response writer that captures the status code and body
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.body == nil {
		rw.body = bytes.NewBuffer(b)
	} else {
		rw.body.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware logs information about each request
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture the status code and body
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           bytes.NewBuffer(nil),
			}

			// Capture request body if debug level is enabled
			var requestBody []byte
			if zerolog.GlobalLevel() == zerolog.DebugLevel {
				if r.Body != nil {
					requestBody, _ = io.ReadAll(r.Body)
					// Restore the request body for the handler
					r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				}
			}

			// Process request
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start)

			// Get logger
			log := logger.Get()

			// Create event based on status code
			event := log.Info()
			if rw.statusCode >= 400 {
				event = log.Error()
			}

			// Log request details
			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("ip", getClientIP(r)).
				Int("status", rw.statusCode).
				Dur("duration", duration).
				Str("user_agent", r.UserAgent())

			// Add request/response bodies in debug mode
			if zerolog.GlobalLevel() == zerolog.DebugLevel {
				if len(requestBody) > 0 {
					event.RawJSON("request_body", requestBody)
				}
				if rw.body.Len() > 0 {
					event.RawJSON("response_body", rw.body.Bytes())
				}
			}

			event.Msg("request completed")
		})
	}
}
