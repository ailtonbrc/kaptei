package handlers

import (
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type LeadHandler struct {
	service ports.LeadService
}

func NewLeadHandler(s ports.LeadService) *LeadHandler {
	return &LeadHandler{service: s}
}

func (h *LeadHandler) ProcessarWebhook(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token") // Usando roteamento do Go 1.22
	h.processarWebhookComToken(w, r, token)
}

func (h *LeadHandler) ProcessarWebhookSeguro(w http.ResponseWriter, r *http.Request) {
	h.processarWebhookComToken(w, r, r.Header.Get("X-Kaptei-Token"))
}

func (h *LeadHandler) processarWebhookComToken(w http.ResponseWriter, r *http.Request, token string) {
	if token == "" {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "token nÃ£o fornecido"})
		return
	}

	var captura domain.CapturaLeadWebhook
	if err := decodificarJSONLimitado(w, r, &captura, 32*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "JSON invÃ¡lido"})
		return
	}

	err := h.service.ProcessarWebhook(r.Context(), token, captura)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}

	responderJSON(w, http.StatusCreated, map[string]string{"mensagem": "lead recebido com sucesso"})
}

func (h *LeadHandler) List(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	leads, err := h.service.List(r.Context(), usuario.ContaID, usuario.ID, usuario.Papel, filtroPaginacao(r))
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, leads)
}

func (h *LeadHandler) Atribuir(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	leadID := r.PathValue("id")

	var req struct {
		UsuarioID string `json:"usuario_id"`
	}
	if err := decodificarJSONLimitado(w, r, &req, 16*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Dados invÃ¡lidos")
		return
	}

	if err := h.service.Atribuir(r.Context(), leadID, usuario.ContaID, usuario.ID, req.UsuarioID, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "lead atribuÃ­do com sucesso"})
}

func (h *LeadHandler) Qualificar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	leadID := r.PathValue("id")

	if err := h.service.Qualificar(r.Context(), leadID, usuario.ContaID, usuario.ID, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "lead qualificado com sucesso"})
}

func (h *LeadHandler) Descartar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	leadID := r.PathValue("id")

	var req struct {
		Motivo string `json:"motivo"`
	}
	if err := decodificarJSONLimitado(w, r, &req, 16*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Dados invÃ¡lidos")
		return
	}

	if err := h.service.Descartar(r.Context(), leadID, usuario.ContaID, usuario.ID, req.Motivo, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "lead descartado com sucesso"})
}
