package gateways

import (
	"context"
	"errors"

	"github.com/msdev/kaptei/internal/core/ports"
)

type PicPayGateway struct{ picpayToken, sellerToken string }

func NewPicPayGateway(token, sellerToken string) ports.PaymentGateway {
	return &PicPayGateway{picpayToken: token, sellerToken: sellerToken}
}

func (g *PicPayGateway) CreateCheckout(context.Context, ports.CheckoutAssinatura) (string, error) {
	if g.picpayToken == "" || g.sellerToken == "" {
		return "", errors.New("gateway PicPay não configurado")
	}
	return "", errors.New("checkout PicPay indisponível até a conclusão da homologação")
}
func (g *PicPayGateway) CreateCustomerPortal(context.Context, string, string) (string, error) {
	return "", errors.New("portal de assinatura PicPay não homologado")
}
func (g *PicPayGateway) ParseWebhook([]byte, string) (ports.EventoPagamento, error) {
	return ports.EventoPagamento{}, errors.New("webhook PicPay não homologado")
}
func (g *PicPayGateway) CancelSubscription(context.Context, string) error {
	return errors.New("cancelamento PicPay não homologado")
}
