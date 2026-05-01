package manager

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TargetRepo interface {
	CreateTarget(ctx context.Context, url string, inverval, timeout int) (string, error)
	DeleteTarget(ctx context.Context, id string) error
}

type Handler struct {
	repo TargetRepo
}

func NewHandler(repo TargetRepo) *Handler {
	return &Handler{repo: repo}
}

type CreateTargetReq struct {
	URL             string `json:"url"`
	IntervalSeconds int    `json:"inverval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
}

func (h *Handler) CreateTarget(w http.ResponseWriter, r *http.Request) {
	var req CreateTargetReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.IntervalSeconds == 0 {
		req.IntervalSeconds = 60
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 5
	}

	id, err := h.repo.CreateTarget(r.Context(), req.URL, req.IntervalSeconds, req.TimeoutSeconds)
	if err != nil {
		http.Error(w, "Failed to create target", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) DeleteTarget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteTarget(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete target", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
