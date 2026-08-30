package handlers

import (
	"encoding/json"
	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"net/http"
)

type DashboardHandler struct {
	service ports.DashboardService
}

func NewDashboardHandler(service ports.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetResumoPremium(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	contaID := user.ContaID

	data, err := h.service.GetDashboardPremium(r.Context(), contaID, user.ID, user.Papel)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "Erro ao buscar dashboard")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
