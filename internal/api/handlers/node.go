package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FelixWinchester/CodeCartographer/internal/db"
	"github.com/go-chi/chi/v5"
)

type NodeHandler struct {
	graphRepo   *db.GraphRepository
	metricsRepo *db.MetricsRepository
}

func NewNodeHandler(graphRepo *db.GraphRepository, metricsRepo *db.MetricsRepository) *NodeHandler {
	return &NodeHandler{
		graphRepo:   graphRepo,
		metricsRepo: metricsRepo,
	}
}

// GetNode возвращает детали узла + метрики
// GET /api/node/{id}
func (h *NodeHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid node id", http.StatusBadRequest)
		return
	}

	node, err := h.graphRepo.GetNode(r.Context(), id)
	if err != nil {
		http.Error(w, "node not found", http.StatusNotFound)
		return
	}

	// Получаем метрики по файлу узла
	metrics, err := h.metricsRepo.GetFileMetrics(r.Context(), node.RepoPath, node.FilePath)
	if err != nil {
		// Метрики опциональны — не падаем если нет
		metrics = nil
	}

	resp := map[string]any{
		"node":    node,
		"metrics": metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
