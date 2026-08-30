package ports

import "context"

type BillingService interface {
	GeneratePaymentLink(ctx context.Context, usuarioID, plano, idempotencyKey string) (string, error)
	GeneratePortalLink(ctx context.Context, usuarioID string) (string, error)
	ProcessWebhook(ctx context.Context, payload []byte, assinatura string) error
}

type BillingRepository interface {
	ProcessarEvento(ctx context.Context, evento EventoPagamento) (bool, error)
}
