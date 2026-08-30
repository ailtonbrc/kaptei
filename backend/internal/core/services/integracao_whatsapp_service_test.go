package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
)

type repositorioWhatsAppTeste struct {
	configuracao *domain.ConfiguracaoWhatsApp
	eventos      []*domain.EventoIntegracao
}

func (r *repositorioWhatsAppTeste) ObterPorConta(context.Context, string) (*domain.ConfiguracaoWhatsApp, error) {
	return r.configuracao, nil
}
func (r *repositorioWhatsAppTeste) ObterPorNumeroTelefone(context.Context, string) (*domain.ConfiguracaoWhatsApp, error) {
	return r.configuracao, nil
}
func (r *repositorioWhatsAppTeste) Salvar(_ context.Context, configuracao *domain.ConfiguracaoWhatsApp) error {
	r.configuracao = configuracao
	return nil
}
func (r *repositorioWhatsAppTeste) Enfileirar(_ context.Context, eventos []*domain.EventoIntegracao) error {
	r.eventos = append(r.eventos, eventos...)
	return nil
}
func (*repositorioWhatsAppTeste) Reservar(context.Context, string, int, time.Duration) ([]*domain.EventoIntegracao, error) {
	return nil, nil
}
func (*repositorioWhatsAppTeste) RegistrarMensagem(context.Context, string, *domain.MensagemWhatsAppRecebida, string, string) error {
	return nil
}
func (*repositorioWhatsAppTeste) Concluir(context.Context, string, string) error { return nil }
func (*repositorioWhatsAppTeste) Falhar(context.Context, string, string, string, time.Time, bool) error {
	return nil
}
func (*repositorioWhatsAppTeste) ObterConversa(context.Context, string, string) (*domain.ConversaWhatsApp, error) {
	return nil, nil
}
func (*repositorioWhatsAppTeste) ListarConversas(context.Context, string, *string, domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.ConversaWhatsApp], error) {
	return &domain.ListaPaginada[*domain.ConversaWhatsApp]{}, nil
}
func (*repositorioWhatsAppTeste) ListarMensagens(context.Context, string, string, domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.MensagemWhatsApp], error) {
	return &domain.ListaPaginada[*domain.MensagemWhatsApp]{}, nil
}
func (*repositorioWhatsAppTeste) CriarMensagemSaida(context.Context, *domain.SolicitacaoEnvioWhatsApp, string, *domain.EventoOutbox) error {
	return nil
}
func (*repositorioWhatsAppTeste) MarcarMensagemEnviada(context.Context, string, string) error {
	return nil
}
func (*repositorioWhatsAppTeste) ObterIdentificadorMensagemSaida(context.Context, string) (*string, error) {
	return nil, nil
}
func (*repositorioWhatsAppTeste) AtualizarStatusMensagem(context.Context, string, string, string, string, time.Time) error {
	return nil
}
func (*repositorioWhatsAppTeste) RegistrarConsentimento(context.Context, string, string, bool, string, string) error {
	return nil
}

type preparadorWhatsAppTeste struct{}

func (preparadorWhatsAppTeste) PrepararMensagem(contaID string, mensagem domain.MensagemWhatsAppRecebida) (*domain.EventoIntegracao, error) {
	return &domain.EventoIntegracao{ContaID: contaID, Provedor: domain.ProvedorWhatsApp,
		Tipo: domain.TipoEventoWhatsAppMensagemEntrada, IdentificadorExterno: mensagem.IdentificadorExterno,
		PaginaID: mensagem.NumeroTelefoneID, PayloadProtegido: "protegido"}, nil
}
func (preparadorWhatsAppTeste) DecodificarMensagem(*domain.EventoIntegracao) (*domain.MensagemWhatsAppRecebida, error) {
	return nil, nil
}
func (preparadorWhatsAppTeste) DecodificarConteudo(string) (*domain.MensagemWhatsAppRecebida, error) {
	return &domain.MensagemWhatsAppRecebida{}, nil
}

func TestReceberWebhookWhatsAppEnfileiraMensagemAssinada(t *testing.T) {
	repo := &repositorioWhatsAppTeste{configuracao: &domain.ConfiguracaoWhatsApp{ContaID: "conta-1", NumeroTelefoneID: "123456", Ativa: true}}
	servico := NewIntegracaoWhatsAppService(repo, protetorMetaTeste{}, "segredo-aplicativo", "token-verificacao-com-mais-de-32-caracteres", preparadorWhatsAppTeste{}, nil, nil)
	corpo := []byte(`{"object":"whatsapp_business_account","entry":[{"changes":[{"field":"messages","value":{"messaging_product":"whatsapp","metadata":{"phone_number_id":"123456"},"contacts":[{"wa_id":"5565999999999","profile":{"name":"Maria"}}],"messages":[{"from":"5565999999999","id":"wamid.exemplo","timestamp":"1720000000","type":"text","text":{"body":"Quero conhecer o imóvel"}}]}}]}]}`)
	mac := hmac.New(sha256.New, []byte("segredo-aplicativo"))
	_, _ = mac.Write(corpo)
	assinatura := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if err := servico.ReceberWebhook(context.Background(), assinatura, corpo); err != nil {
		t.Fatalf("webhook WhatsApp válido foi recusado: %v", err)
	}
	if len(repo.eventos) != 1 || repo.eventos[0].IdentificadorExterno != "wamid.exemplo" || repo.eventos[0].ContaID != "conta-1" {
		t.Fatalf("evento WhatsApp não foi enfileirado corretamente: %+v", repo.eventos)
	}
}

func TestReceberWebhookWhatsAppRejeitaAssinaturaInvalida(t *testing.T) {
	repo := &repositorioWhatsAppTeste{}
	servico := NewIntegracaoWhatsAppService(repo, protetorMetaTeste{}, "segredo-aplicativo", "token-verificacao-com-mais-de-32-caracteres", preparadorWhatsAppTeste{}, nil, nil)
	if err := servico.ReceberWebhook(context.Background(), "sha256=00", []byte(`{}`)); err == nil {
		t.Fatal("assinatura inválida deveria ser rejeitada")
	}
}
