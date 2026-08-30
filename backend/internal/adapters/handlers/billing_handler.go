package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type BillingHandler struct{ billingService ports.BillingService }

var chaveIdempotenciaValida = regexp.MustCompile(`^[A-Za-z0-9_.:-]{8,255}$`)

func NewBillingHandler(servico ports.BillingService) *BillingHandler {
	return &BillingHandler{billingService: servico}
}

func (h *BillingHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario == nil {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	if usuario.Papel != domain.RoleGestor && usuario.Papel != domain.RoleCorretorSolo {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": "apenas o responsável pela conta pode alterar a assinatura"})
		return
	}
	var requisicao struct {
		Plano string `json:"plano"`
	}
	if err := decodificarJSONLimitado(w, r, &requisicao, 8*1024); err != nil || strings.TrimSpace(requisicao.Plano) == "" {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "plano é obrigatório"})
		return
	}
	idempotencia := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if !chaveIdempotenciaValida.MatchString(idempotencia) {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "Idempotency-Key válida é obrigatória"})
		return
	}
	link, err := h.billingService.GeneratePaymentLink(r.Context(), usuario.ID, requisicao.Plano, idempotencia)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"checkout_url": link})
}

func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "payload inválido"})
		return
	}
	if err := h.billingService.ProcessWebhook(r.Context(), payload, r.Header.Get("Stripe-Signature")); err != nil {
		slog.ErrorContext(r.Context(), "falha ao processar webhook Stripe", "erro", err)
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "evento de cobrança inválido"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BillingHandler) Portal(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario == nil {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	if usuario.Papel != domain.RoleGestor && usuario.Papel != domain.RoleCorretorSolo {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": "apenas o responsável pela conta pode gerenciar a assinatura"})
		return
	}
	link, err := h.billingService.GeneratePortalLink(r.Context(), usuario.ID)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"portal_url": link})
}
