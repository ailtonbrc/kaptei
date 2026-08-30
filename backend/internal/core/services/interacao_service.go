package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type interacaoService struct {
	repo        ports.InteracaoRepository
	clienteRepo ports.ClienteRepository
}

func NewInteracaoService(repo ports.InteracaoRepository, clienteRepo ports.ClienteRepository) ports.InteracaoService {
	return &interacaoService{repo: repo, clienteRepo: clienteRepo}
}

func (s *interacaoService) Create(ctx context.Context, interacao *domain.Interacao, usuarioAtorID string, papel domain.Role) error {
	if interacao == nil || interacao.ContaID == "" || interacao.ClienteID == "" {
		return errors.New("interação, conta e cliente são obrigatórios")
	}
	if err := s.validarAcessoCliente(ctx, interacao.ClienteID, interacao.ContaID, usuarioAtorID, papel); err != nil {
		return err
	}
	interacao.Tipo = strings.ToUpper(strings.TrimSpace(interacao.Tipo))
	interacao.Descricao = strings.TrimSpace(interacao.Descricao)
	tiposValidos := map[string]bool{"LIGACAO": true, "MENSAGEM": true, "VISITA": true, "PROPOSTA": true, "ANOTACAO": true}
	if !tiposValidos[interacao.Tipo] || interacao.Descricao == "" || len([]rune(interacao.Descricao)) > 4000 {
		return errors.New("tipo válido e descrição de até 4.000 caracteres são obrigatórios")
	}
	interacao.CorretorID = &usuarioAtorID
	if interacao.DataHora.IsZero() {
		interacao.DataHora = time.Now()
	}
	if interacao.DataHora.Year() < 1900 || interacao.DataHora.After(time.Now().Add(5*time.Minute)) {
		return errors.New("data da interação inválida")
	}
	return s.repo.Create(ctx, interacao)
}

func (s *interacaoService) ListByClienteID(ctx context.Context, clienteID, contaID, usuarioAtorID string, papel domain.Role) ([]*domain.Interacao, error) {
	if err := s.validarAcessoCliente(ctx, clienteID, contaID, usuarioAtorID, papel); err != nil {
		return nil, err
	}
	return s.repo.ListByClienteID(ctx, clienteID, contaID)
}

func (s *interacaoService) Delete(ctx context.Context, id, contaID, usuarioAtorID string, papel domain.Role) error {
	interacao, err := s.repo.GetByID(ctx, id, contaID)
	if err != nil || interacao == nil {
		return errors.New("interação não encontrada")
	}
	if !podeGerenciarClientes(papel) && (interacao.CorretorID == nil || *interacao.CorretorID != usuarioAtorID) {
		return errors.New("sem permissão para excluir esta interação")
	}
	return s.repo.Delete(ctx, id, contaID)
}

func (s *interacaoService) validarAcessoCliente(ctx context.Context, clienteID, contaID, usuarioAtorID string, papel domain.Role) error {
	cliente, err := s.clienteRepo.GetByID(ctx, clienteID, contaID)
	if err != nil || cliente == nil {
		return errors.New("cliente não encontrado")
	}
	if !podeGerenciarClientes(papel) && (cliente.CorretorID == nil || *cliente.CorretorID != usuarioAtorID) {
		return errors.New("sem permissão para acessar este cliente")
	}
	return nil
}
