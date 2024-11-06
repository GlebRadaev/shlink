package handlers

import (
	"net/http"

	"github.com/GlebRadaev/shlink/internal/service"
)

type HealthHandlers struct {
	healthService *service.HealthService
}

func NewHealthHandlers(healthService *service.HealthService) *HealthHandlers {
	return &HealthHandlers{healthService: healthService}
}

func (h *HealthHandlers) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.healthService.CheckDatabaseConnection(r.Context()); err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
