package gateways

import (
	"context"
	"net"
)

type resolvedorDNS struct{ resolvedor *net.Resolver }

func NewResolvedorDNS() *resolvedorDNS { return &resolvedorDNS{resolvedor: net.DefaultResolver} }

func (r *resolvedorDNS) ConsultarTXT(ctx context.Context, nome string) ([]string, error) {
	return r.resolvedor.LookupTXT(ctx, nome)
}
