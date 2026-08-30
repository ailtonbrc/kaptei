package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
)

type repositorioMetaTeste struct {
	porConta  *domain.ConfiguracaoMetaLeads
	porPagina *domain.ConfiguracaoMetaLeads
	eventos   []*domain.EventoIntegracao
}

func (r *repositorioMetaTeste) ObterPorConta(context.Context, string) (*domain.ConfiguracaoMetaLeads, error) {
	return r.porConta, nil
}
func (r *repositorioMetaTeste) ObterPorPagina(context.Context, string) (*domain.ConfiguracaoMetaLeads, error) {
	return r.porPagina, nil
}
func (r *repositorioMetaTeste) Salvar(_ context.Context, configuracao *domain.ConfiguracaoMetaLeads) error {
	r.porConta = configuracao
	return nil
}
func (r *repositorioMetaTeste) Enfileirar(_ context.Context, eventos []*domain.EventoIntegracao) error {
	r.eventos = append(r.eventos, eventos...)
	return nil
}
func (*repositorioMetaTeste) Reservar(context.Context, string, int, time.Duration) ([]*domain.EventoIntegracao, error) {
	return nil, nil
}
func (*repositorioMetaTeste) Concluir(context.Context, string, string) error { return nil }
func (*repositorioMetaTeste) Falhar(context.Context, string, string, string, time.Time, bool) error {
	return nil
}

type protetorMetaTeste struct{}

func (protetorMetaTeste) Proteger(valor string) (string, error) { return "protegido:" + valor, nil }
func (protetorMetaTeste) Revelar(valor string) (string, error) {
	if len(valor) <= len("protegido:") {
		return "", errors.New("invÃ¡lido")
	}
	return valor[len("protegido:"):], nil
}

func TestValidarAssinaturaMeta(t *testing.T) {
	corpo := []byte(`{"object":"page","entry":[]}`)
	mac := hmac.New(sha256.New, []byte("segredo-aplicativo"))
	_, _ = mac.Write(corpo)
	assinatura := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !validarAssinaturaMeta(corpo, assinatura, "segredo-aplicativo") {
		t.Fatal("assinatura correta deveria ser aceita")
	}
	if validarAssinaturaMeta([]byte("alterado"), assinatura, "segredo-aplicativo") {
		t.Fatal("corpo alterado nÃ£o pode ser aceito")
	}
}

func TestReceberWebhookMetaMapeiaEventoConhecido(t *testing.T) {
	repo := &repositorioMetaTeste{porPagina: &domain.ConfiguracaoMetaLeads{ContaID: "conta-1", PaginaID: "123", Ativa: true}}
	servico := NewIntegracaoMetaService(repo, protetorMetaTeste{}, "segredo-aplicativo", "token-verificacao-com-mais-de-32-caracteres", 8)
	corpo := []byte(`{"object":"page","entry":[{"id":"123","changes":[{"field":"leadgen","value":{"leadgen_id":"456","page_id":"123","form_id":"789","ad_id":"987","created_time":1720000000}}]}]}`)
	mac := hmac.New(sha256.New, []byte("segredo-aplicativo"))
	_, _ = mac.Write(corpo)
	assinatura := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if err := servico.ReceberWebhook(context.Background(), assinatura, corpo); err != nil {
		t.Fatalf("webhook vÃ¡lido foi recusado: %v", err)
	}
	if len(repo.eventos) != 1 || repo.eventos[0].IdentificadorExterno != "456" || repo.eventos[0].ContaID != "conta-1" {
		t.Fatalf("evento nÃ£o foi mapeado corretamente: %+v", repo.eventos)
	}
}

func TestSalvarMetaNaoDevolveToken(t *testing.T) {
	repo := &repositorioMetaTeste{}
	servico := NewIntegracaoMetaService(repo, protetorMetaTeste{}, "segredo-aplicativo", "token-verificacao-com-mais-de-32-caracteres", 8)
	configuracao, err := servico.SalvarConfiguracao(context.Background(), "conta-1", domain.RoleGestor, domain.AtualizacaoMetaLeads{
		PaginaID: "123456", TokenPagina: "token-da-pagina-com-tamanho-valido", Ativa: true,
	})
	if err != nil {
		t.Fatalf("configuraÃ§Ã£o vÃ¡lida foi recusada: %v", err)
	}
	if configuracao.TokenPaginaProtegido != "" || !configuracao.TokenPaginaConfigurado {
		t.Fatalf("segredo foi exposto ou nÃ£o marcado: %+v", configuracao)
	}
}

func TestUUIDMetaDeterministico(t *testing.T) {
	primeiro := uuidDeterministico("meta:123")
	segundo := uuidDeterministico("meta:123")
	if primeiro != segundo || !formatoUUID.MatchString(primeiro) {
		t.Fatalf("UUID idempotente invÃ¡lido: %q e %q", primeiro, segundo)
	}
}
