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

type IntegracaoWhatsAppHandler struct {
	servico ports.IntegracaoWhatsAppService
}

func (h *IntegracaoWhatsAppHandler) VerificarWebhook(w http.ResponseWriter, r *http.Request) {
	desafio, err := h.servico.VerificarWebhook(r.URL.Query().Get("hub.mode"), r.URL.Query().Get("hub.verify_token"), r.URL.Query().Get("hub.challenge"))
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

func (h *IntegracaoWhatsAppHandler) ReceberWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, limiteWebhookMeta)
	corpo, err := io.ReadAll(r.Body)
	if err != nil {
		responderJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"erro": "payload WhatsApp excede o limite permitido"})
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

func NewIntegracaoWhatsAppHandler(servico ports.IntegracaoWhatsAppService) *IntegracaoWhatsAppHandler {
	return &IntegracaoWhatsAppHandler{servico: servico}
}

func (h *IntegracaoWhatsAppHandler) ObterConfiguracao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	configuracao, err := h.servico.ObterConfiguracao(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, configuracao)
}

func (h *IntegracaoWhatsAppHandler) SalvarConfiguracao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	var atualizacao domain.AtualizacaoWhatsApp
	if err := decodificarJSONLimitado(w, r, &atualizacao, 16*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados da integração WhatsApp inválidos"})
		return
	}
	configuracao, err := h.servico.SalvarConfiguracao(r.Context(), usuario.ContaID, usuario.Papel, atualizacao)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, configuracao)
}

func (h *IntegracaoWhatsAppHandler) ListarConversas(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	resultado, err := h.servico.ListarConversas(r.Context(), usuario.ContaID, usuario.ID, usuario.Papel, filtroPaginacao(r))
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, resultado)
}

func (h *IntegracaoWhatsAppHandler) ListarMensagens(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	resultado, err := h.servico.ListarMensagens(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, usuario.Papel, filtroPaginacao(r))
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, resultado)
}

func (h *IntegracaoWhatsAppHandler) EnviarTexto(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	var requisicao struct {
		Texto string `json:"texto"`
	}
	if err := decodificarJSONLimitado(w, r, &requisicao, 16*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "mensagem inválida"})
		return
	}
	if err := h.servico.EnviarTexto(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, usuario.Papel, requisicao.Texto); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusAccepted, map[string]string{"status": "enfileirada"})
}

func (h *IntegracaoWhatsAppHandler) EnviarTemplate(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	var template domain.EnvioTemplateWhatsApp
	if err := decodificarJSONLimitado(w, r, &template, 32*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "template inválido"})
		return
	}
	if err := h.servico.EnviarTemplate(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, usuario.Papel, template); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusAccepted, map[string]string{"status": "enfileirada"})
}

func (h *IntegracaoWhatsAppHandler) RegistrarConsentimento(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	var requisicao struct {
		Consentiu bool   `json:"consentiu"`
		Origem    string `json:"origem"`
		Evidencia string `json:"evidencia"`
	}
	if err := decodificarJSONLimitado(w, r, &requisicao, 8*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "consentimento inválido"})
		return
	}
	if err := h.servico.RegistrarConsentimento(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, usuario.Papel, requisicao.Consentiu, requisicao.Origem, requisicao.Evidencia); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"status": "atualizado"})
}
