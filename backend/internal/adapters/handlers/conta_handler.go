package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type ContaHandler struct {
	servico ports.ContaService
}

func NewContaHandler(servico ports.ContaService) *ContaHandler {
	return &ContaHandler{servico: servico}
}

func (h *ContaHandler) GetMinhaConta(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	conta, err := h.servico.Obter(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil || conta == nil {
		responderErro(w, http.StatusNotFound, "Conta nÃ£o encontrada")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conta)
}

func (h *ContaHandler) UpdateLeadConfig(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	var req struct {
		Estrategia string `json:"lead_estrategia"`
	}
	if err := decodificarJSONLimitado(w, r, &req, 8*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "JSON invÃ¡lido")
		return
	}

	if err := h.servico.AtualizarEstrategiaLeads(r.Context(), usuario.ContaID, req.Estrategia, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	conta, err := h.servico.Obter(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "nÃ£o foi possÃ­vel carregar a conta atualizada"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conta)
}

func (h *ContaHandler) RotacionarTokenLeads(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "nÃ£o autorizado"})
		return
	}
	token, err := h.servico.RotacionarTokenLeads(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{
		"token": token, "prefixo": token[:8],
		"aviso": "copie agora; o token completo nÃ£o serÃ¡ exibido novamente",
	})
}
