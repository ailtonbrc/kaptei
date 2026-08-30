package ports

import (
	"context"
	"time"
)

type CheckoutAssinatura struct {
	ContaID, PlanoCodigo, PriceID, Email, CustomerID, URLSucesso, URLCancelamento, IdempotencyKey string
}

type EventoPagamento struct {
	ID, Tipo, ContaID, CustomerID, SubscriptionID, PlanoCodigo, Status string
	PeriodoFim                                                         *time.Time
}

type PaymentGateway interface {
	CreateCheckout(ctx context.Context, checkout CheckoutAssinatura) (string, error)
	CreateCustomerPortal(ctx context.Context, customerID, returnURL string) (string, error)
	ParseWebhook(payload []byte, assinatura string) (EventoPagamento, error)
	CancelSubscription(ctx context.Context, subscriptionID string) error
}
