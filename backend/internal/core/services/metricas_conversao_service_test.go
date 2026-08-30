package services

import (
	"context"
	"errors"
	"testing"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type siteMetricasConversaoFalso struct {
	ports.SitePublicoRepository
	site   *domain.SitePublico
	imovel *domain.ImovelPublico
}

func (r *siteMetricasConversaoFalso) GetBySlug(context.Context, string) (*domain.SitePublico, error) {
	return r.site, nil
}

func (r *siteMetricasConversaoFalso) GetImovelBySlug(context.Context, string, string) (*domain.ImovelPublico, error) {
	return r.imovel, nil
}

type repositorioMetricasConversaoFalso struct {
	ports.MetricasConversaoRepository
	evento *domain.EventoConversao
}

func (r *repositorioMetricasConversaoFalso) Registrar(_ context.Context, evento *domain.EventoConversao) error {
	r.evento = evento
	return nil
}

func TestMetricasConversaoRecusaIdentificadoresInvalidosAntesDeConsultarSite(t *testing.T) {
	sites := &siteMetricasConversaoFalso{}
	repositorio := &repositorioMetricasConversaoFalso{}
	servico := NewMetricasConversaoService(sites, repositorio)
	err := servico.Registrar(context.Background(), "site", domain.EventoConversaoPublico{ChaveEvento: "invalida", SessaoID: "invalida", Tipo: domain.EventoSiteVisualizado})
	if !errors.Is(err, ErrEventoConversaoInvalido) || repositorio.evento != nil {
		t.Fatal("evento inválido não deve alcançar o repositório")
	}
}

func TestMetricasConversaoVinculaImovelDoMesmoSite(t *testing.T) {
	sites := &siteMetricasConversaoFalso{
		site:   &domain.SitePublico{ContaID: "conta-1", Slug: "site"},
		imovel: &domain.ImovelPublico{ID: "00000000-0000-4000-8000-000000000003", Slug: "imovel"},
	}
	repositorio := &repositorioMetricasConversaoFalso{}
	servico := NewMetricasConversaoService(sites, repositorio)
	imovelSlug := " imovel "
	fonte := " campanha "
	err := servico.Registrar(context.Background(), "site", domain.EventoConversaoPublico{
		ChaveEvento: "00000000-0000-4000-8000-000000000001", SessaoID: "00000000-0000-4000-8000-000000000002",
		Tipo: domain.EventoImovelVisualizado, ImovelSlug: &imovelSlug, UTMSource: &fonte,
	})
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if repositorio.evento == nil || repositorio.evento.ImovelID == nil || *repositorio.evento.ImovelID != sites.imovel.ID {
		t.Fatal("imóvel validado não foi vinculado ao evento")
	}
	if repositorio.evento.UTMSource == nil || *repositorio.evento.UTMSource != "campanha" {
		t.Fatal("atribuição não foi normalizada")
	}
}
