package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/FelixWinchester/CodeCartographer/internal/db"
)

type GraphHandler struct {
	graphRepo *db.GraphRepository
}

func NewGraphHandler(graphRepo *db.GraphRepository) *GraphHandler {
	return &GraphHandler{graphRepo: graphRepo}
}

// GetGraph возвращает весь граф зависимостей
// GET /api/graph?repo=<repo_path>
func (h *GraphHandler) GetGraph(w http.ResponseWriter, r *http.Request) {
	repoPath := r.URL.Query().Get("repo")
	if repoPath == "" {
		http.Error(w, "repo query parameter is required", http.StatusBadRequest)
		return
	}

	graph, err := h.graphRepo.GetGraph(r.Context(), repoPath)
	if err != nil {
		http.Error(w, "failed to get graph", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}
