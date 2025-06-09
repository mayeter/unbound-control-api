package handler

import (
	"encoding/json"
	"net/http"

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

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (h *UnboundHandler) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.client.Status()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    status,
	})
}

func (h *UnboundHandler) Reload(w http.ResponseWriter, r *http.Request) {
	response, err := h.client.Reload()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

func (h *UnboundHandler) Flush(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		respondWithError(w, http.StatusBadRequest, "Domain is required")
		return
	}
	response, err := h.client.Flush(domain)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

func (h *UnboundHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.client.Stats()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, Response{
		Success: false,
		Error:   message,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
