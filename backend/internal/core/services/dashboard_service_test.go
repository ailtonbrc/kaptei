package services

import (
	"context"
	"testing"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
)

type dashboardRepositorioFalso struct {
	ultimoUsuarioID *string
}

func (r *dashboardRepositorioFalso) GetFunilConversao(_ context.Context, _ string, usuarioID *string) (map[string]int, error) {
	r.ultimoUsuarioID = usuarioID
	return map[string]int{"NOVO": 3}, nil
}

func (r *dashboardRepositorioFalso) GetOrigemLeads(context.Context, string, *string) (map[string]int, error) {
	return map[string]int{"SITE": 3}, nil
}

func (r *dashboardRepositorioFalso) GetMetricasResumo(context.Context, string, *string) (map[string]interface{}, error) {
	return map[string]interface{}{"total_imoveis": 4}, nil
}

func (r *dashboardRepositorioFalso) GetEvolucaoLeads(context.Context, string, *string) ([]string, []int, error) {
	return []string{"2026-08"}, []int{3}, nil
}

type metricasConversaoRepositorioFalso struct {
	consultas int
}

func (*metricasConversaoRepositorioFalso) Registrar(context.Context, *domain.EventoConversao) error {
	return nil
}

func (r *metricasConversaoRepositorioFalso) ObterResumo(context.Context, string, time.Time) (*domain.ResumoConversaoSite, error) {
	r.consultas++
	return &domain.ResumoConversaoSite{VisitasSite: 10, ContatosEnviados: 2, TaxaConversao: 20}, nil
}

func (*metricasConversaoRepositorioFalso) ExpurgarExpirados(context.Context) (int64, error) {
	return 0, nil
}

func TestDashboardIncluiMetricasConversaoParaGestor(t *testing.T) {
	repositorio := &dashboardRepositorioFalso{}
	metricas := &metricasConversaoRepositorioFalso{}
	servico := NewDashboardService(repositorio, metricas)

	resposta, err := servico.GetDashboardPremium(context.Background(), "conta-1", "gestor-1", domain.RoleGestor)
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if resposta["conversao_site"] == nil {
		t.Fatal("dashboard do gestor deve incluir as métricas de conversão do site")
	}
	if metricas.consultas != 1 {
		t.Fatalf("esperava uma consulta de conversão, recebeu %d", metricas.consultas)
	}
	if repositorio.ultimoUsuarioID != nil {
		t.Fatal("gestor deve visualizar o resumo da conta inteira")
	}
}

func TestDashboardNaoExpoeMetricasGlobaisAoCorretorEquipe(t *testing.T) {
	repositorio := &dashboardRepositorioFalso{}
	metricas := &metricasConversaoRepositorioFalso{}
	servico := NewDashboardService(repositorio, metricas)

	resposta, err := servico.GetDashboardPremium(context.Background(), "conta-1", "corretor-1", domain.RoleCorretorEquipe)
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if _, existe := resposta["conversao_site"]; existe {
		t.Fatal("corretor de equipe não deve receber métricas globais da imobiliária")
	}
	if metricas.consultas != 0 {
		t.Fatal("métricas globais não devem ser consultadas para corretor de equipe")
	}
	if repositorio.ultimoUsuarioID == nil || *repositorio.ultimoUsuarioID != "corretor-1" {
		t.Fatal("consultas do corretor devem permanecer filtradas pelo usuário")
	}
}
