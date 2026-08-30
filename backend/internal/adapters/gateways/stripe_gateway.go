package gateways

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/ports"
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
)

type StripeGateway struct {
	cliente        *stripe.Client
	chave          string
	segredoWebhook string
}

func NewStripeGateway(chave, segredoWebhook string) ports.PaymentGateway {
	return &StripeGateway{cliente: stripe.NewClient(chave), chave: chave, segredoWebhook: segredoWebhook}
}

func (g *StripeGateway) CreateCheckout(ctx context.Context, checkout ports.CheckoutAssinatura) (string, error) {
	if g.chave == "" {
		return "", errors.New("Stripe não configurado")
	}
	if checkout.PriceID == "" {
		return "", errors.New("plano sem preço configurado no gateway")
	}
	params := &stripe.CheckoutSessionCreateParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ClientReferenceID: stripe.String(checkout.ContaID),
		SuccessURL:        stripe.String(checkout.URLSucesso), CancelURL: stripe.String(checkout.URLCancelamento),
		LineItems:        []*stripe.CheckoutSessionCreateLineItemParams{{Price: stripe.String(checkout.PriceID), Quantity: stripe.Int64(1)}},
		Metadata:         map[string]string{"conta_id": checkout.ContaID, "plano": checkout.PlanoCodigo},
		SubscriptionData: &stripe.CheckoutSessionCreateSubscriptionDataParams{Metadata: map[string]string{"conta_id": checkout.ContaID, "plano": checkout.PlanoCodigo}},
	}
	if checkout.CustomerID != "" {
		params.Customer = stripe.String(checkout.CustomerID)
	} else {
		params.CustomerEmail = stripe.String(checkout.Email)
	}
	if checkout.IdempotencyKey != "" {
		params.SetIdempotencyKey(checkout.IdempotencyKey)
	}
	sessao, err := g.cliente.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return "", fmt.Errorf("criar checkout Stripe: %w", err)
	}
	if sessao.URL == "" {
		return "", errors.New("Stripe não retornou a URL do checkout")
	}
	return sessao.URL, nil
}

func (g *StripeGateway) ParseWebhook(payload []byte, assinatura string) (ports.EventoPagamento, error) {
	if g.segredoWebhook == "" {
		return ports.EventoPagamento{}, errors.New("segredo do webhook Stripe não configurado")
	}
	evento, err := webhook.ConstructEvent(payload, assinatura, g.segredoWebhook)
	if err != nil {
		return ports.EventoPagamento{}, errors.New("assinatura do webhook inválida")
	}
	var objeto map[string]any
	if err := json.Unmarshal(evento.Data.Raw, &objeto); err != nil {
		return ports.EventoPagamento{}, errors.New("evento Stripe inválido")
	}
	resultado := ports.EventoPagamento{ID: evento.ID, Tipo: string(evento.Type)}
	resultado.ContaID = primeiroTexto(objeto, "client_reference_id", "metadata.conta_id", "parent.subscription_details.metadata.conta_id")
	resultado.PlanoCodigo = primeiroTexto(objeto, "metadata.plano", "parent.subscription_details.metadata.plano")
	resultado.CustomerID = textoObjeto(objeto["customer"])
	resultado.SubscriptionID = primeiroTexto(objeto, "subscription")
	if strings.HasPrefix(resultado.Tipo, "customer.subscription.") {
		resultado.SubscriptionID = primeiroTexto(objeto, "id")
	}
	resultado.Status = primeiroTexto(objeto, "status", "payment_status")
	if timestamp := primeiroNumero(objeto, "current_period_end", "period_end"); timestamp > 0 {
		periodoFim := time.Unix(timestamp, 0).UTC()
		resultado.PeriodoFim = &periodoFim
	}
	if resultado.Tipo == "invoice.paid" || resultado.Tipo == "invoice.payment_failed" {
		resultado.SubscriptionID = primeiroTexto(objeto, "subscription", "parent.subscription_details.subscription")
	}
	return resultado, nil
}

func primeiroNumero(objeto map[string]any, caminhos ...string) int64 {
	for _, caminho := range caminhos {
		atual := any(objeto)
		for _, parte := range strings.Split(caminho, ".") {
			mapa, ok := atual.(map[string]any)
			if !ok {
				atual = nil
				break
			}
			atual = mapa[parte]
		}
		if numero, ok := atual.(float64); ok {
			return int64(numero)
		}
	}
	return 0
}

func (g *StripeGateway) CreateCustomerPortal(ctx context.Context, customerID, returnURL string) (string, error) {
	if g.chave == "" {
		return "", errors.New("Stripe não configurado")
	}
	if customerID == "" {
		return "", errors.New("cliente ainda não vinculado ao gateway")
	}
	sessao, err := g.cliente.V1BillingPortalSessions.Create(ctx, &stripe.BillingPortalSessionCreateParams{
		Customer: stripe.String(customerID), ReturnURL: stripe.String(returnURL),
	})
	if err != nil {
		return "", fmt.Errorf("criar portal do cliente Stripe: %w", err)
	}
	if sessao.URL == "" {
		return "", errors.New("Stripe não retornou a URL do portal")
	}
	return sessao.URL, nil
}

func (g *StripeGateway) CancelSubscription(ctx context.Context, subscriptionID string) error {
	if g.chave == "" {
		return errors.New("Stripe não configurado")
	}
	_, err := g.cliente.V1Subscriptions.Cancel(ctx, subscriptionID, nil)
	return err
}

func primeiroTexto(objeto map[string]any, caminhos ...string) string {
	for _, caminho := range caminhos {
		atual := any(objeto)
		for _, parte := range strings.Split(caminho, ".") {
			mapa, ok := atual.(map[string]any)
			if !ok {
				atual = nil
				break
			}
			atual = mapa[parte]
		}
		if texto := textoObjeto(atual); texto != "" {
			return texto
		}
	}
	return ""
}

func textoObjeto(valor any) string {
	if texto, ok := valor.(string); ok {
		return texto
	}
	if objeto, ok := valor.(map[string]any); ok {
		if id, ok := objeto["id"].(string); ok {
			return id
		}
	}
	return ""
}
