package services

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

const confirmacaoExecucaoRetencao = "ANONIMIZAR DADOS EXPIRADOS"

type retencaoService struct{ repo ports.RetencaoRepository }

func NewRetencaoService(repo ports.RetencaoRepository) ports.RetencaoService {
	return &retencaoService{repo: repo}
}

func (s *retencaoService) ObterPolitica(ctx context.Context, contaID string, papel domain.Role) (*domain.PoliticaRetencao, error) {
	if !podeGerenciarPrivacidade(papel) {
		return nil, errors.New("sem permissão para gerenciar retenção")
	}
	return s.repo.ObterPolitica(ctx, contaID)
}

func (s *retencaoService) SalvarPolitica(ctx context.Context, contaID, usuarioID string, papel domain.Role, politica domain.PoliticaRetencao) error {
	if !podeGerenciarPrivacidade(papel) {
		return errors.New("sem permissão para gerenciar retenção")
	}
	politica.ContaID = contaID
	politica.FundamentoLegal = strings.TrimSpace(politica.FundamentoLegal)
	if politica.DiasLeadsDescartados < 30 || politica.DiasLeadsDescartados > 3650 || politica.DiasClientesPerdidos < 30 || politica.DiasClientesPerdidos > 3650 {
		return errors.New("prazos de retenção devem estar entre 30 e 3.650 dias")
	}
	if politica.TamanhoLote < 1 || politica.TamanhoLote > 1000 {
		return errors.New("lote deve possuir entre 1 e 1.000 registros por tipo")
	}
	if utf8.RuneCountInString(politica.FundamentoLegal) > 2_000 || (politica.Ativa && utf8.RuneCountInString(politica.FundamentoLegal) < 10) {
		return errors.New("informe o fundamento legal da política ativa")
	}
	return s.repo.SalvarPolitica(ctx, &politica, usuarioID)
}

func (s *retencaoService) GerarRelatorio(ctx context.Context, contaID string, papel domain.Role) (*domain.RelatorioRetencao, error) {
	if !podeGerenciarPrivacidade(papel) {
		return nil, errors.New("sem permissão para consultar retenção")
	}
	politica, err := s.repo.ObterPolitica(ctx, contaID)
	if err != nil {
		return nil, err
	}
	return s.repo.GerarRelatorio(ctx, contaID, politica)
}

func (s *retencaoService) Executar(ctx context.Context, contaID, usuarioID string, papel domain.Role, confirmacao string) (*domain.ResultadoRetencao, error) {
	if !podeGerenciarPrivacidade(papel) {
		return nil, errors.New("sem permissão para executar retenção")
	}
	if confirmacao != confirmacaoExecucaoRetencao {
		return nil, errors.New("confirmação de retenção inválida")
	}
	politica, err := s.repo.ObterPolitica(ctx, contaID)
	if err != nil {
		return nil, err
	}
	if !politica.Ativa {
		return nil, errors.New("política de retenção está desativada")
	}
	return s.repo.Executar(ctx, contaID, usuarioID, politica)
}

func (s *retencaoService) ListarBloqueios(ctx context.Context, contaID string, papel domain.Role) ([]domain.BloqueioRetencao, error) {
	if !podeGerenciarPrivacidade(papel) {
		return nil, errors.New("sem permissão para consultar bloqueios")
	}
	return s.repo.ListarBloqueios(ctx, contaID)
}

func (s *retencaoService) SalvarBloqueio(ctx context.Context, contaID, usuarioID string, papel domain.Role, bloqueio domain.BloqueioRetencao) (*domain.BloqueioRetencao, error) {
	if !podeGerenciarPrivacidade(papel) {
		return nil, errors.New("sem permissão para gerenciar bloqueios")
	}
	bloqueio.ContaID = contaID
	bloqueio.TipoRecurso = strings.ToUpper(strings.TrimSpace(bloqueio.TipoRecurso))
	bloqueio.Motivo = strings.TrimSpace(bloqueio.Motivo)
	if bloqueio.TipoRecurso != "LEAD" && bloqueio.TipoRecurso != "CLIENTE" {
		return nil, errors.New("tipo de recurso inválido")
	}
	if !formatoUUID.MatchString(bloqueio.RecursoID) {
		return nil, errors.New("identificador do recurso inválido")
	}
	if utf8.RuneCountInString(bloqueio.Motivo) < 5 || utf8.RuneCountInString(bloqueio.Motivo) > 1_000 {
		return nil, errors.New("motivo deve ter entre 5 e 1.000 caracteres")
	}
	if bloqueio.ValidoAte != nil && !bloqueio.ValidoAte.After(time.Now()) {
		return nil, errors.New("vigência do bloqueio deve estar no futuro")
	}
	if err := s.repo.SalvarBloqueio(ctx, &bloqueio, usuarioID); err != nil {
		return nil, err
	}
	return &bloqueio, nil
}

func (s *retencaoService) RemoverBloqueio(ctx context.Context, id, contaID string, papel domain.Role) error {
	if !podeGerenciarPrivacidade(papel) {
		return errors.New("sem permissão para remover bloqueios")
	}
	if !formatoUUID.MatchString(id) {
		return errors.New("identificador do bloqueio inválido")
	}
	return s.repo.RemoverBloqueio(ctx, id, contaID)
}
