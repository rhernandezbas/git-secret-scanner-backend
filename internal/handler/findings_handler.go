package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

type FindingsHandler struct {
	repo domain.FindingsRepository
}

func NewFindingsHandler(repo domain.FindingsRepository) *FindingsHandler {
	return &FindingsHandler{repo: repo}
}

func (h *FindingsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	results, err := h.repo.LoadAll(context.Background())
	if err != nil {
		http.Error(w, `{"error":"failed to load findings"}`, http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []domain.ScanResult{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
