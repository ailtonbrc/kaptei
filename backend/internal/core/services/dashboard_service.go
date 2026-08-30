package services

import (
	"context"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type dashboardService struct {
	repo              ports.DashboardRepository
	metricasConversao ports.MetricasConversaoRepository
}

func NewDashboardService(repo ports.DashboardRepository, metricasConversao ports.MetricasConversaoRepository) ports.DashboardService {
	return &dashboardService{repo: repo, metricasConversao: metricasConversao}
}

func (s *dashboardService) GetDashboardPremium(ctx context.Context, contaID, usuarioAtorID string, papel domain.Role) (map[string]interface{}, error) {
	var usuarioID *string
	if papel != domain.RoleGestor && papel != domain.RoleSuperAdmin {
		usuarioID = &usuarioAtorID
	}
	funil, err := s.repo.GetFunilConversao(ctx, contaID, usuarioID)
	if err != nil {
		return nil, err
	}

	origens, err := s.repo.GetOrigemLeads(ctx, contaID, usuarioID)
	if err != nil {
		return nil, err
	}

	resumo, err := s.repo.GetMetricasResumo(ctx, contaID, usuarioID)
	if err != nil {
		return nil, err
	}
	categorias, valores, err := s.repo.GetEvolucaoLeads(ctx, contaID, usuarioID)
	if err != nil {
		return nil, err
	}

	// Monta a resposta base antes de incluir blocos opcionais.
	response := map[string]interface{}{
		"metricas":       resumo,
		"funil":          funil,
		"origens":        origens,
		"leads_evolucao": map[string]interface{}{"categorias": categorias, "valores": valores},
	}
	if papel != domain.RoleCorretorEquipe && s.metricasConversao != nil {
		conversao, err := s.metricasConversao.ObterResumo(ctx, contaID, time.Now().UTC().AddDate(0, 0, -30))
		if err != nil {
			return nil, err
		}
		response["conversao_site"] = conversao
	}

	return response, nil
}
