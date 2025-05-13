package handler

import (
	"encoding/json"
	"net/http"

	"github.com/callMe-Root/unbound-control-api/internal/unbound"
	"github.com/callMe-Root/unbound-control-api/internal/zonefile"
	"github.com/gorilla/mux"
)

type ZoneFileHandler struct {
	client *unbound.Client
}

func NewZoneFileHandler(client *unbound.Client) *ZoneFileHandler {
	return &ZoneFileHandler{
		client: client,
	}
}

// GetZoneFile handles GET /api/v1/zones/{name}/file
func (h *ZoneFileHandler) GetZoneFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]

	zoneFile, err := h.client.GetZoneFile(zoneName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    zoneFile,
	})
}

// UpdateZoneFile handles PUT /api/v1/zones/{name}/file
func (h *ZoneFileHandler) UpdateZoneFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]

	var zoneFile zonefile.ZoneFile
	if err := json.NewDecoder(r.Body).Decode(&zoneFile); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Ensure the zone name in the URL matches the zone name in the payload
	if zoneFile.Name != zoneName {
		respondWithError(w, http.StatusBadRequest, "Zone name mismatch")
		return
	}

	if err := h.client.UpdateZoneFile(zoneName, &zoneFile); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    zoneFile,
	})
}

// AddZoneRecord handles POST /api/v1/zones/{name}/records
func (h *ZoneFileHandler) AddZoneRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]

	var record zonefile.Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.client.AddZoneRecord(zoneName, record); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    record,
	})
}

// GetZoneRecord handles GET /api/v1/zones/{name}/records/{recordName}/{recordType}
func (h *ZoneFileHandler) GetZoneRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]
	recordName := vars["recordName"]
	recordType := vars["recordType"]

	record, err := h.client.GetZoneRecord(zoneName, recordName, recordType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    record,
	})
}

// UpdateZoneRecord handles PUT /api/v1/zones/{name}/records/{recordName}/{recordType}
func (h *ZoneFileHandler) UpdateZoneRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]
	recordName := vars["recordName"]
	recordType := vars["recordType"]

	var record zonefile.Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Ensure the record name and type in the URL match the record in the payload
	if record.Name != recordName || record.Type != recordType {
		respondWithError(w, http.StatusBadRequest, "Record name or type mismatch")
		return
	}

	if err := h.client.UpdateZoneRecord(zoneName, record); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    record,
	})
}

// RemoveZoneRecord handles DELETE /api/v1/zones/{name}/records/{recordName}/{recordType}
func (h *ZoneFileHandler) RemoveZoneRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName := vars["name"]
	recordName := vars["recordName"]
	recordType := vars["recordType"]

	if err := h.client.RemoveZoneRecord(zoneName, recordName, recordType); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    "Record removed successfully",
	})
}
