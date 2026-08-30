package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/msdev/kaptei/internal/core/ports"
)

type billingService struct {
	contaRepo   ports.ContaRepository
	userRepo    ports.UsuarioRepository
	planoRepo   ports.PlanoRepository
	gateway     ports.PaymentGateway
	billingRepo ports.BillingRepository
	urlPublica  string
}

func NewBillingService(c ports.ContaRepository, u ports.UsuarioRepository, p ports.PlanoRepository, g ports.PaymentGateway, b ports.BillingRepository, urlPublica string) ports.BillingService {
	return &billingService{contaRepo: c, userRepo: u, planoRepo: p, gateway: g, billingRepo: b, urlPublica: strings.TrimRight(urlPublica, "/")}
}

func (s *billingService) GeneratePaymentLink(ctx context.Context, usuarioID, planoCodigo, idempotencyKey string) (string, error) {
	if s.urlPublica == "" {
		return "", errors.New("URL pública da aplicação não configurada")
	}
	usuario, err := s.userRepo.GetByID(ctx, usuarioID)
	if err != nil || usuario == nil {
		return "", errors.New("usuário não encontrado")
	}
	conta, err := s.contaRepo.GetByID(ctx, usuario.ContaID)
	if err != nil || conta == nil {
		return "", errors.New("conta não encontrada")
	}
	plano, err := s.planoRepo.GetByCodigo(ctx, strings.TrimSpace(planoCodigo))
	if err != nil || plano == nil || !plano.Ativo {
		return "", errors.New("plano inválido")
	}
	tipoEsperado := "CORRETOR"
	if conta.TipoConta == "IMOBILIARIA" {
		tipoEsperado = "IMOBILIARIA"
	}
	if plano.Tipo != tipoEsperado {
		return "", errors.New("plano incompatível com o tipo da conta")
	}
	if plano.GatewayPriceID == nil || *plano.GatewayPriceID == "" {
		return "", errors.New("plano ainda não configurado para cobrança")
	}
	if conta.BillingSubscriptionID != nil && *conta.BillingSubscriptionID != "" && conta.StatusPlano != "CANCELADO" {
		return "", errors.New("já existe uma assinatura vinculada; use o portal para gerenciá-la")
	}
	customerID := ""
	if conta.BillingCustomerID != nil {
		customerID = *conta.BillingCustomerID
	}
	return s.gateway.CreateCheckout(ctx, ports.CheckoutAssinatura{
		ContaID: conta.ID, PlanoCodigo: plano.Codigo, PriceID: *plano.GatewayPriceID,
		Email: usuario.Email, CustomerID: customerID, IdempotencyKey: idempotencyKey,
		URLSucesso:      s.urlPublica + "/checkout/sucesso?session_id={CHECKOUT_SESSION_ID}",
		URLCancelamento: s.urlPublica + "/app/assinatura",
	})
}

func (s *billingService) ProcessWebhook(ctx context.Context, payload []byte, assinatura string) error {
	evento, err := s.gateway.ParseWebhook(payload, assinatura)
	if err != nil {
		return err
	}
	if evento.ID == "" || evento.Tipo == "" {
		return errors.New("evento de cobrança incompleto")
	}
	_, err = s.billingRepo.ProcessarEvento(ctx, evento)
	if err != nil {
		return fmt.Errorf("processar evento de cobrança: %w", err)
	}
	return nil
}

func (s *billingService) GeneratePortalLink(ctx context.Context, usuarioID string) (string, error) {
	if s.urlPublica == "" {
		return "", errors.New("URL pública da aplicação não configurada")
	}
	usuario, err := s.userRepo.GetByID(ctx, usuarioID)
	if err != nil || usuario == nil {
		return "", errors.New("usuário não encontrado")
	}
	conta, err := s.contaRepo.GetByID(ctx, usuario.ContaID)
	if err != nil || conta == nil || conta.BillingCustomerID == nil || *conta.BillingCustomerID == "" {
		return "", errors.New("assinatura ainda não vinculada ao gateway")
	}
	return s.gateway.CreateCustomerPortal(ctx, *conta.BillingCustomerID, s.urlPublica+"/app/assinatura")
}
