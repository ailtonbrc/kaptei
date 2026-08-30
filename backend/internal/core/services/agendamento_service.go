package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type agendamentoService struct {
	repo        ports.AgendamentoRepository
	clienteRepo ports.ClienteRepository
	usuarioRepo ports.UsuarioRepository
	imovelRepo  ports.ImovelRepository
}

func NewAgendamentoService(repo ports.AgendamentoRepository, clienteRepo ports.ClienteRepository, usuarioRepo ports.UsuarioRepository, imovelRepo ports.ImovelRepository) ports.AgendamentoService {
	return &agendamentoService{repo: repo, clienteRepo: clienteRepo, usuarioRepo: usuarioRepo, imovelRepo: imovelRepo}
}

func (s *agendamentoService) Create(ctx context.Context, agendamento *domain.Agendamento, usuarioAtorID string, papel domain.Role) error {
	if papel != domain.RoleGestor && papel != domain.RoleSuperAdmin {
		agendamento.UsuarioID = usuarioAtorID
	} else {
		responsavel, err := s.usuarioRepo.GetByID(ctx, agendamento.UsuarioID)
		if err != nil || responsavel == nil || responsavel.ContaID != agendamento.ContaID || !strings.EqualFold(responsavel.Status, "ATIVO") {
			return errors.New("responsável pelo agendamento é inválido")
		}
		if responsavel.Papel != domain.RoleGestor && responsavel.Papel != domain.RoleCorretorEquipe && responsavel.Papel != domain.RoleCorretorSolo {
			return errors.New("usuário selecionado não pode receber agendamentos")
		}
	}
	if agendamento.ImovelID != nil {
		imovel, err := s.imovelRepo.GetByID(ctx, *agendamento.ImovelID, agendamento.ContaID)
		if err != nil || imovel == nil {
			return errors.New("imóvel do agendamento não encontrado")
		}
	}
	if agendamento.ClienteID != nil {
		cliente, err := s.clienteRepo.GetByID(ctx, *agendamento.ClienteID, agendamento.ContaID)
		if err != nil || cliente == nil {
			return errors.New("cliente do agendamento não encontrado")
		}
		if papel != domain.RoleGestor && papel != domain.RoleSuperAdmin && (cliente.CorretorID == nil || *cliente.CorretorID != usuarioAtorID) {
			return errors.New("sem permissão para agendar para este cliente")
		}
	}
	if agendamento.Status == "" {
		agendamento.Status = domain.StatusAgendamentoAgendado
	}
	if agendamento.Tipo == "" {
		agendamento.Tipo = domain.TipoAgendamentoVisita
	}
	if err := validarAgendamento(agendamento); err != nil {
		return err
	}
	return s.repo.Create(ctx, agendamento)
}

func (s *agendamentoService) GetByID(ctx context.Context, id, contaID string) (*domain.Agendamento, error) {
	if id == "" || contaID == "" {
		return nil, errors.New("agendamento e conta são obrigatórios")
	}
	return s.repo.GetByID(ctx, id, contaID)
}

func (s *agendamentoService) List(ctx context.Context, contaID string, usuarioID *string, inicio, fim time.Time) ([]*domain.Agendamento, error) {
	if contaID == "" || inicio.IsZero() || fim.IsZero() || !fim.After(inicio) {
		return nil, errors.New("período de consulta inválido")
	}
	return s.repo.List(ctx, contaID, usuarioID, inicio, fim)
}

func (s *agendamentoService) Update(ctx context.Context, agendamento *domain.Agendamento) error {
	if agendamento.ID == "" {
		return errors.New("agendamento é obrigatório")
	}
	if err := validarAgendamento(agendamento); err != nil {
		return err
	}
	return s.repo.Update(ctx, agendamento)
}

func (s *agendamentoService) Delete(ctx context.Context, id, contaID string) error {
	if id == "" || contaID == "" {
		return errors.New("agendamento e conta são obrigatórios")
	}
	return s.repo.Delete(ctx, id, contaID)
}

func validarAgendamento(agendamento *domain.Agendamento) error {
	if agendamento == nil || agendamento.ContaID == "" || agendamento.UsuarioID == "" {
		return errors.New("conta e usuário são obrigatórios")
	}
	agendamento.Titulo = strings.TrimSpace(agendamento.Titulo)
	if agendamento.Titulo == "" {
		return errors.New("título é obrigatório")
	}
	if len([]rune(agendamento.Titulo)) > 255 {
		return errors.New("título excede 255 caracteres")
	}
	agendamento.Descricao = strings.TrimSpace(agendamento.Descricao)
	if len([]rune(agendamento.Descricao)) > 4000 {
		return errors.New("descrição excede 4.000 caracteres")
	}
	if agendamento.DataHoraInicio.IsZero() || agendamento.DataHoraFim.IsZero() || !agendamento.DataHoraFim.After(agendamento.DataHoraInicio) {
		return errors.New("data final deve ser posterior à data inicial")
	}
	if agendamento.DataHoraFim.Sub(agendamento.DataHoraInicio) > 24*time.Hour {
		return errors.New("um agendamento não pode exceder 24 horas")
	}
	statusValidos := map[domain.StatusAgendamento]bool{
		domain.StatusAgendamentoAgendado: true, domain.StatusAgendamentoConfirmado: true,
		domain.StatusAgendamentoConcluido: true, domain.StatusAgendamentoCancelado: true,
		domain.StatusAgendamentoNaoCompareceu: true,
	}
	if !statusValidos[agendamento.Status] {
		return errors.New("status do agendamento inválido")
	}
	tiposValidos := map[domain.TipoAgendamento]bool{
		domain.TipoAgendamentoVisita: true, domain.TipoAgendamentoLigacao: true,
		domain.TipoAgendamentoReuniaoOnline: true, domain.TipoAgendamentoTarefa: true,
	}
	if !tiposValidos[agendamento.Tipo] {
		return errors.New("tipo do agendamento inválido")
	}
	return nil
}
