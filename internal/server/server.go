package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/callMe-Root/unbound-control-api/internal/config"
	"github.com/callMe-Root/unbound-control-api/internal/middleware"
	"github.com/callMe-Root/unbound-control-api/internal/unbound"
	"github.com/gorilla/mux"
)

// Server represents our HTTP server
type Server struct {
	httpServer *http.Server
	router     *mux.Router
	certFile   string
	keyFile    string
	mu         sync.RWMutex
	config     *config.Config
	client     *unbound.Client
}

// New creates a new server instance
func New(host string, port int, certFile, keyFile string, cfg *config.Config, client *unbound.Client) *Server {
	router := mux.NewRouter()
	addr := fmt.Sprintf("%s:%d", host, port)

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		router:   router,
		certFile: certFile,
		keyFile:  keyFile,
		config:   cfg,
		client:   client,
	}
}

// Router returns the server's router
func (s *Server) Router() *mux.Router {
	return s.router
}

// reloadConfig reloads the configuration and updates the server accordingly
func (s *Server) reloadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load new configuration
	newCfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load new configuration: %w", err)
	}

	// Update API key authentication
	s.router.Use(middleware.APIKeyAuth(newCfg.Security.APIKey))

	// Update rate limiting
	s.router.Use(middleware.RateLimit(newCfg.RateLimit.RequestsPerSecond, newCfg.RateLimit.BurstSize))

	// Update Unbound client settings if changed
	if newCfg.Unbound.ControlHost != s.config.Unbound.ControlHost ||
		newCfg.Unbound.ControlPort != s.config.Unbound.ControlPort ||
		newCfg.Unbound.ControlCert != s.config.Unbound.ControlCert ||
		newCfg.Unbound.ControlKey != s.config.Unbound.ControlKey {

		// Create new client with updated settings
		newClient, err := unbound.NewClient(
			newCfg.Unbound.ControlHost,
			newCfg.Unbound.ControlPort,
			newCfg.Unbound.ControlCert,
			newCfg.Unbound.ControlKey,
		)
		if err != nil {
			return fmt.Errorf("failed to create new Unbound client: %w", err)
		}

		// Close old client and update to new one
		s.client.Close()
		s.client = newClient
	}

	// Update server configuration
	s.config = newCfg

	return nil
}

// reloadCertificate reloads the TLS certificate and key
func (s *Server) reloadCertificate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load the new certificate
	cert, err := tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	// Get the current TLS config
	tlsConfig := s.httpServer.TLSConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}

	// Update the certificate
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Update the server's TLS config
	s.httpServer.TLSConfig = tlsConfig

	return nil
}

// Start starts the server with TLS support
func (s *Server) Start() error {
	// Create a channel to listen for errors coming from the server
	serverErrors := make(chan error, 1)

	// Create a channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// Start the server
	go func() {
		if s.certFile != "" && s.keyFile != "" {
			log.Printf("Starting server with TLS on %s", s.httpServer.Addr)
			serverErrors <- s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile)
		} else {
			log.Printf("Starting server without TLS on %s", s.httpServer.Addr)
			serverErrors <- s.httpServer.ListenAndServe()
		}
	}()

	// Blocking main and waiting for shutdown
	for {
		select {
		case err := <-serverErrors:
			return fmt.Errorf("server error: %w", err)

		case sig := <-shutdown:
			switch sig {
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, reloading configuration and certificates")

				// Reload configuration
				if err := s.reloadConfig(); err != nil {
					log.Printf("Failed to reload configuration: %v", err)
				} else {
					log.Println("Configuration reloaded successfully")
				}

				// Reload certificate if TLS is enabled
				if s.certFile != "" && s.keyFile != "" {
					if err := s.reloadCertificate(); err != nil {
						log.Printf("Failed to reload certificate: %v", err)
					} else {
						log.Println("Certificate reloaded successfully")
					}
				}

			case os.Interrupt, syscall.SIGTERM:
				log.Println("Shutting down server...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := s.httpServer.Shutdown(ctx); err != nil {
					return fmt.Errorf("could not stop server gracefully: %w", err)
				}
				return nil
			}
		}
	}
}
