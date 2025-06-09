package main

import (
	"log"

	"github.com/callMe-Root/unbound-control-api/internal/config"
	"github.com/callMe-Root/unbound-control-api/internal/handler"
	"github.com/callMe-Root/unbound-control-api/internal/middleware"
	"github.com/callMe-Root/unbound-control-api/internal/server"
	"github.com/callMe-Root/unbound-control-api/internal/unbound"
	"github.com/callMe-Root/unbound-control-api/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger.Initialize(cfg.Logging.Level, cfg.Logging.UseSyslog, cfg.Logging.AppName)

	// Create Unbound client
	client, err := unbound.NewClient(cfg.Unbound.ControlSocket)
	if err != nil {
		log.Fatalf("Failed to create Unbound client: %v", err)
	}
	defer client.Close()

	// Create server
	var certFile, keyFile string
	if cfg.Server.UseTLS {
		certFile = cfg.Server.CertFile
		keyFile = cfg.Server.KeyFile
	}
	srv := server.New(cfg.Server.Host, cfg.Server.Port, certFile, keyFile, cfg, client)

	// Add logging middleware
	srv.Router().Use(middleware.LoggingMiddleware())

	// Create handlers
	unboundHandler := handler.NewUnboundHandler(client)
	zoneHandler := handler.NewZoneHandler(client)
	zoneFileHandler := handler.NewZoneFileHandler(client)

	// API routes with authentication and rate limiting
	api := srv.Router().PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.APIKeyAuth(cfg.Security.APIKey))
	api.Use(middleware.RateLimit(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.BurstSize))

	// Unbound control routes
	api.HandleFunc("/status", unboundHandler.Status).Methods("GET")
	api.HandleFunc("/reload", unboundHandler.Reload).Methods("POST")
	api.HandleFunc("/flush", unboundHandler.Flush).Methods("DELETE")
	api.HandleFunc("/stats", unboundHandler.Stats).Methods("GET")

	// Zone management routes
	api.HandleFunc("/zones", zoneHandler.ListZones).Methods("GET")
	api.HandleFunc("/zones", zoneHandler.AddZone).Methods("POST")
	api.HandleFunc("/zones/{name}", zoneHandler.GetZone).Methods("GET")
	api.HandleFunc("/zones/{name}", zoneHandler.UpdateZone).Methods("PUT")
	api.HandleFunc("/zones/{name}", zoneHandler.RemoveZone).Methods("DELETE")

	// Zone file management routes
	api.HandleFunc("/zones/{name}/file", zoneFileHandler.GetZoneFile).Methods("GET")
	api.HandleFunc("/zones/{name}/file", zoneFileHandler.UpdateZoneFile).Methods("PUT")
	api.HandleFunc("/zones/{name}/records", zoneFileHandler.AddZoneRecord).Methods("POST")
	api.HandleFunc("/zones/{name}/records/{recordName}/{recordType}", zoneFileHandler.GetZoneRecord).Methods("GET")
	api.HandleFunc("/zones/{name}/records/{recordName}/{recordType}", zoneFileHandler.UpdateZoneRecord).Methods("PUT")
	api.HandleFunc("/zones/{name}/records/{recordName}/{recordType}", zoneFileHandler.RemoveZoneRecord).Methods("DELETE")

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
