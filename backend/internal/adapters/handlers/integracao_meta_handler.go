package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"github.com/msdev/kaptei/internal/core/services"
)

const limiteWebhookMeta = 1024 * 1024

type IntegracaoMetaHandler struct {
	servico ports.IntegracaoMetaService
}

func NewIntegracaoMetaHandler(servico ports.IntegracaoMetaService) *IntegracaoMetaHandler {
	return &IntegracaoMetaHandler{servico: servico}
}

func (h *IntegracaoMetaHandler) VerificarWebhook(w http.ResponseWriter, r *http.Request) {
	desafio, err := h.servico.VerificarWebhook(
		r.URL.Query().Get("hub.mode"),
		r.URL.Query().Get("hub.verify_token"),
		r.URL.Query().Get("hub.challenge"),
	)
	if err != nil {
		status := http.StatusForbidden
		if errors.Is(err, services.ErrMetaIndisponivel) {
			status = http.StatusServiceUnavailable
		}
		responderJSON(w, status, map[string]string{"erro": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(desafio))
}

func (h *IntegracaoMetaHandler) ReceberWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, limiteWebhookMeta)
	corpo, err := io.ReadAll(r.Body)
	if err != nil {
		responderJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"erro": "payload Meta excede o limite permitido"})
		return
	}
	if err := h.servico.ReceberWebhook(r.Context(), r.Header.Get("X-Hub-Signature-256"), corpo); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrMetaAssinaturaInvalida) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, services.ErrMetaIndisponivel) {
			status = http.StatusServiceUnavailable
		}
		responderJSON(w, status, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"status": "recebido"})
}

func (h *IntegracaoMetaHandler) ObterConfiguracao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "nÃ£o autorizado"})
		return
	}
	configuracao, err := h.servico.ObterConfiguracao(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, configuracao)
}

func (h *IntegracaoMetaHandler) SalvarConfiguracao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "nÃ£o autorizado"})
		return
	}
	var atualizacao domain.AtualizacaoMetaLeads
	if err := decodificarJSONLimitado(w, r, &atualizacao, 16*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados da integraÃ§Ã£o Meta invÃ¡lidos"})
		return
	}
	configuracao, err := h.servico.SalvarConfiguracao(r.Context(), usuario.ContaID, usuario.Papel, atualizacao)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, configuracao)
}
