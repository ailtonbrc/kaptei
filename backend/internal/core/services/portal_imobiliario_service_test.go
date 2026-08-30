package services

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type repositorioPortalWebhookFalso struct {
	ports.PortalImobiliarioRepository
	consultasToken int
	imovelID       string
}

func (r *repositorioPortalWebhookFalso) ObterContaPorToken(context.Context, string, string) (string, error) {
	r.consultasToken++
	return "conta-1", nil
}

func (r *repositorioPortalWebhookFalso) ObterImovelDaConta(context.Context, string, string) (*string, error) {
	return &r.imovelID, nil
}

type leadServiceWebhookFalso struct {
	ports.LeadService
	capturas []domain.CapturaLeadIntegracao
}

func (s *leadServiceWebhookFalso) CaptarIntegracao(_ context.Context, _ string, captura domain.CapturaLeadIntegracao) error {
	s.capturas = append(s.capturas, captura)
	return nil
}

func autorizacaoTeste(segredo string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte("vivareal:"+segredo))
}

func leadGrupoOLXValido() domain.LeadGrupoOLX {
	return domain.LeadGrupoOLX{
		LeadOrigin: "Grupo OLX", Timestamp: "2026-08-07T12:00:00Z", OriginLeadID: "lead-externo-1",
		ClientListingID: "00000000-0000-4000-8000-000000000001", Name: "Pessoa Interessada",
		Email: "pessoa@example.com", DDD: "65", Phone: "999999999", Message: "Tenho interesse no imóvel.",
		Temperature: "Alta", TransactionType: "SELL", ExtraData: domain.ExtraDataLeadGrupoOLX{LeadType: "CONTACT_FORM"},
	}
}

func TestWebhookGrupoOLXValidaAutorizacaoAntesDoTenant(t *testing.T) {
	repositorio := &repositorioPortalWebhookFalso{imovelID: "00000000-0000-4000-8000-000000000001"}
	leads := &leadServiceWebhookFalso{}
	servico := NewPortalImobiliarioService(repositorio, nil, leads, "https://app.exemplo.com.br", "segredo-global-seguro")

	err := servico.ReceberLead(context.Background(), string(make([]byte, 64)), "Basic inválido", leadGrupoOLXValido())
	if err == nil {
		t.Fatal("era esperado erro de autenticação")
	}
	if repositorio.consultasToken != 0 || len(leads.capturas) != 0 {
		t.Fatal("requisição não autenticada não deve consultar tenant nem criar lead")
	}
}

func TestWebhookGrupoOLXMapeiaImovelEChaveIdempotente(t *testing.T) {
	imovelID := "00000000-0000-4000-8000-000000000001"
	repositorio := &repositorioPortalWebhookFalso{imovelID: imovelID}
	leads := &leadServiceWebhookFalso{}
	servico := NewPortalImobiliarioService(repositorio, nil, leads, "https://app.exemplo.com.br", "segredo-global-seguro")
	token := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	if err := servico.ReceberLead(context.Background(), token, autorizacaoTeste("segredo-global-seguro"), leadGrupoOLXValido()); err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if len(leads.capturas) != 1 {
		t.Fatalf("capturas inesperadas: %d", len(leads.capturas))
	}
	captura := leads.capturas[0]
	if captura.ImovelID == nil || *captura.ImovelID != imovelID {
		t.Fatal("lead não foi vinculado ao imóvel validado")
	}
	if captura.ChaveIdempotencia != uuidDeterministico("grupo-olx:lead-externo-1") {
		t.Fatal("chave idempotente não é determinística")
	}
}
