// Package handlers provides HTTP handlers related to health checks and other
// application management features.
package handlers

import (
	"net/http"

	"github.com/GlebRadaev/shlink/internal/service"
)

// HealthHandlers defines the handlers responsible for application health checks.
type HealthHandlers struct {
	// healthService is the service that manages the health checks.
	healthService *service.HealthService
}

// NewHealthHandlers creates a new instance of HealthHandlers.
func NewHealthHandlers(healthService *service.HealthService) *HealthHandlers {
	return &HealthHandlers{healthService: healthService}
}

// Ping is a handler to check the health of the application.
func (h *HealthHandlers) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.healthService.CheckDatabaseConnection(r.Context()); err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
