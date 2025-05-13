package handler

import (
	"encoding/json"
	"net/http"

	"github.com/callMe-Root/unbound-control-api/internal/unbound"
	"github.com/gorilla/mux"
)

type ZoneHandler struct {
	client *unbound.Client
}

func NewZoneHandler(client *unbound.Client) *ZoneHandler {
	return &ZoneHandler{
		client: client,
	}
}

// ListZones handles GET /api/v1/zones
func (h *ZoneHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	zones, err := h.client.ListZones()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    zones,
	})
}

// GetZone handles GET /api/v1/zones/{name}
func (h *ZoneHandler) GetZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]

	zone, err := h.client.GetZone(zoneName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    zone,
	})
}

// AddZone handles POST /api/v1/zones
func (h *ZoneHandler) AddZone(w http.ResponseWriter, r *http.Request) {
	var zone unbound.Zone
	if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.client.AddZone(zone); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    zone,
	})
}

// UpdateZone handles PUT /api/v1/zones/{name}
func (h *ZoneHandler) UpdateZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]

	var zone unbound.Zone
	if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Ensure the zone name in the URL matches the zone name in the payload
	if zone.Name != zoneName {
		respondWithError(w, http.StatusBadRequest, "Zone name mismatch")
		return
	}

	if err := h.client.UpdateZone(zone); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    zone,
	})
}

// RemoveZone handles DELETE /api/v1/zones/{name}
func (h *ZoneHandler) RemoveZone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]

	if err := h.client.RemoveZone(zoneName); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    "Zone removed successfully",
	})
}
