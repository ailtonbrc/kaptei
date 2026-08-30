package services

import (
	"context"
	"errors"
	"testing"

	"github.com/msdev/kaptei/internal/core/domain"
)

type repositorioDominioFake struct{ dominio *domain.DominioSite }

func (r *repositorioDominioFake) ObterPorConta(context.Context, string) (*domain.DominioSite, error) {
	return r.dominio, nil
}
func (r *repositorioDominioFake) SalvarPendente(_ context.Context, dominio *domain.DominioSite) error {
	dominio.ID = "dominio-1"
	r.dominio = dominio
	return nil
}
func (r *repositorioDominioFake) Ativar(_ context.Context, _, _, token string) error {
	if r.dominio == nil || r.dominio.TokenVerificacao != token {
		return errors.New("versão divergente")
	}
	r.dominio.Status = "ATIVO"
	return nil
}
func (r *repositorioDominioFake) RegistrarFalha(_ context.Context, _, _, token, mensagem string) error {
	if r.dominio == nil || r.dominio.TokenVerificacao != token {
		return errors.New("versão divergente")
	}
	r.dominio.Status = "FALHOU"
	r.dominio.UltimoErro = &mensagem
	return nil
}
func (r *repositorioDominioFake) ObterSitePorHostname(context.Context, string) (*domain.SitePublico, error) {
	return nil, nil
}

type resolvedorDNSFake struct {
	registros []string
	err       error
}

func (r resolvedorDNSFake) ConsultarTXT(context.Context, string) ([]string, error) {
	return r.registros, r.err
}

func TestNormalizarHostname(t *testing.T) {
	hostname, err := normalizarHostname(" Imoveis.Exemplo.COM.BR. ")
	if err != nil || hostname != "imoveis.exemplo.com.br" {
		t.Fatalf("hostname inesperado: %q, erro: %v", hostname, err)
	}
	for _, invalido := range []string{"localhost", "https://exemplo.com", "127.0.0.1", "dominio_sem_tld", "-invalido.com"} {
		if _, err := normalizarHostname(invalido); err == nil {
			t.Fatalf("esperava rejeição para %q", invalido)
		}
	}
}

func TestVerificarDominioExigeTXTExato(t *testing.T) {
	repo := &repositorioDominioFake{dominio: &domain.DominioSite{ID: "dominio-1", ContaID: "conta-1", Hostname: "imoveis.exemplo.com.br", TokenVerificacao: "token-seguro", Status: "PENDENTE"}}
	servico := NewDominioSiteService(repo, resolvedorDNSFake{registros: []string{"kaptei-verificacao=token-seguro"}})
	dominio, err := servico.Verificar(context.Background(), "conta-1", domain.RoleGestor)
	if err != nil {
		t.Fatalf("verificar domínio: %v", err)
	}
	if dominio.Status != "ATIVO" {
		t.Fatalf("status inesperado: %s", dominio.Status)
	}
}

func TestVerificarDominioNaoAceitaTXTDiferente(t *testing.T) {
	repo := &repositorioDominioFake{dominio: &domain.DominioSite{ID: "dominio-1", ContaID: "conta-1", Hostname: "imoveis.exemplo.com.br", TokenVerificacao: "esperado", Status: "PENDENTE"}}
	servico := NewDominioSiteService(repo, resolvedorDNSFake{registros: []string{"kaptei-verificacao=outro"}})
	if _, err := servico.Verificar(context.Background(), "conta-1", domain.RoleGestor); err == nil {
		t.Fatal("esperava falha de verificação")
	}
	if repo.dominio.Status != "FALHOU" {
		t.Fatalf("status inesperado: %s", repo.dominio.Status)
	}
}
