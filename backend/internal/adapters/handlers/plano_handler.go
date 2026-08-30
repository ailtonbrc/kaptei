package handlers

import (
	"encoding/json"
	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"net/http"
	"strings"
)

type PlanoHandler struct {
	planoRepo ports.PlanoRepository
}

func (h *PlanoHandler) ListarAdministracao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario.Papel != domain.RoleSuperAdmin {
		responderErro(w, http.StatusForbidden, "Sem permissÃ£o")
		return
	}
	planos, err := h.planoRepo.ListarAtivos(r.Context())
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "Erro ao buscar planos")
		return
	}
	resposta := make([]map[string]any, 0, len(planos))
	for _, plano := range planos {
		priceID := ""
		if plano.GatewayPriceID != nil {
			priceID = *plano.GatewayPriceID
		}
		resposta = append(resposta, map[string]any{"codigo": plano.Codigo, "nome": plano.Nome, "gateway_price_id": priceID})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resposta)
}

func (h *PlanoHandler) ConfigurarGateway(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario.Papel != domain.RoleSuperAdmin {
		responderErro(w, http.StatusForbidden, "Sem permissÃ£o")
		return
	}
	var req struct {
		PriceID string `json:"gateway_price_id"`
	}
	if err := decodificarJSONLimitado(w, r, &req, 8*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Payload invÃ¡lido")
		return
	}
	priceID := strings.TrimSpace(req.PriceID)
	if priceID != "" && !strings.HasPrefix(priceID, "price_") {
		responderErro(w, http.StatusBadRequest, "Price ID invÃ¡lido")
		return
	}
	if err := h.planoRepo.AtualizarGatewayPriceID(r.Context(), r.PathValue("codigo"), priceID); err != nil {
		responderErro(w, http.StatusNotFound, "Plano nÃ£o encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func NewPlanoHandler(repo ports.PlanoRepository) *PlanoHandler {
	return &PlanoHandler{planoRepo: repo}
}

func (h *PlanoHandler) Listar(w http.ResponseWriter, r *http.Request) {
	planos, err := h.planoRepo.ListarAtivos(r.Context())
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "Erro ao buscar planos")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(planos)
}
