package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/service"
)

type ScanHandler struct {
	svc *service.ScanService
}

func NewScanHandler(svc *service.ScanService) *ScanHandler {
	return &ScanHandler{svc: svc}
}

func (h *ScanHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ScanHandler) Scan(w http.ResponseWriter, r *http.Request) {
	var req domain.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, `{"error":"username is required"}`, http.StatusBadRequest)
		return
	}
	if req.Provider == "" {
		req.Provider = "github"
	}

	// Run scan in background goroutine
	go func() {
		_ = h.svc.Run(r.Context(), req)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "scan started",
		"username": req.Username,
		"provider": req.Provider,
	})
}
