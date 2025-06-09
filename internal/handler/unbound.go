package handler

import (
	"encoding/json"
	"net/http"

	"github.com/callMe-Root/unbound-control-api/internal/response"
	"github.com/callMe-Root/unbound-control-api/internal/unbound"
)

type UnboundHandler struct {
	client *unbound.Client
}

func NewUnboundHandler(client *unbound.Client) *UnboundHandler {
	return &UnboundHandler{
		client: client,
	}
}

func (h *UnboundHandler) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.client.Status()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response.CommonResponse{
		Success: true,
		Data:    status,
	})
}

func (h *UnboundHandler) Reload(w http.ResponseWriter, r *http.Request) {
	err := h.client.Reload()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response.CommonResponse{
		Success: true,
		Data:    "Configuration reloaded successfully",
	})
}

func (h *UnboundHandler) Flush(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		respondWithError(w, http.StatusBadRequest, "Domain is required")
		return
	}

	err := h.client.Flush(domain)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response.CommonResponse{
		Success: true,
		Data:    "Cache flushed successfully",
	})
}

func (h *UnboundHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.client.Stats()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response.CommonResponse{
		Success: true,
		Data:    stats,
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, response.CommonResponse{
		Success: false,
		Error: &response.Error{
			Code:    "INTERNAL_ERROR",
			Message: message,
		},
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
