package services

import (
	"context"
	"errors"
	"strings"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type contaService struct {
	repositorio ports.ContaRepository
}

func NewContaService(repositorio ports.ContaRepository) ports.ContaService {
	return &contaService{repositorio: repositorio}
}

func (s *contaService) Obter(ctx context.Context, contaID string, papel domain.Role) (*domain.ContaSaaS, error) {
	conta, err := s.repositorio.GetByID(ctx, contaID)
	if err != nil {
		return nil, err
	}
	if conta == nil {
		return nil, errors.New("conta não encontrada")
	}
	conta.LeadTokenHash = nil
	if papel == domain.RoleCorretorEquipe {
		conta.LeadTokenIntegracao = nil
		conta.LeadTokenPrefixo = nil
	}
	return conta, nil
}

func (s *contaService) AtualizarEstrategiaLeads(ctx context.Context, contaID, estrategia string, papel domain.Role) error {
	if !podeAdministrarConta(papel) {
		return errors.New("apenas gestores podem alterar configurações de leads")
	}
	estrategia = strings.ToUpper(strings.TrimSpace(estrategia))
	if estrategia != "CAIXA_ENTRADA" && estrategia != "ROLETA" {
		return errors.New("estratégia de leads inválida")
	}
	return s.repositorio.AtualizarEstrategiaLeads(ctx, contaID, estrategia)
}

func (s *contaService) RotacionarTokenLeads(ctx context.Context, contaID string, papel domain.Role) (string, error) {
	if !podeAdministrarConta(papel) {
		return "", errors.New("apenas gestores podem rotacionar o token de leads")
	}
	token, err := generateSecureToken()
	if err != nil {
		return "", err
	}
	if err := s.repositorio.RotacionarTokenLeads(ctx, contaID, hashTokenRecuperacao(token), token[:8]); err != nil {
		return "", err
	}
	return token, nil
}

func podeAdministrarConta(papel domain.Role) bool {
	return papel == domain.RoleSuperAdmin || papel == domain.RoleGestor || papel == domain.RoleCorretorSolo
}
