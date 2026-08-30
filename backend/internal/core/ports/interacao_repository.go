package ports

import (
	"context"
	"github.com/msdev/kaptei/internal/core/domain"
)

type InteracaoRepository interface {
	Create(ctx context.Context, interacao *domain.Interacao) error
	GetByID(ctx context.Context, id, contaID string) (*domain.Interacao, error)
	ListByClienteID(ctx context.Context, clienteID, contaID string) ([]*domain.Interacao, error)
	Delete(ctx context.Context, id, contaID string) error
}

type InteracaoService interface {
	Create(ctx context.Context, interacao *domain.Interacao, usuarioAtorID string, papel domain.Role) error
	ListByClienteID(ctx context.Context, clienteID, contaID, usuarioAtorID string, papel domain.Role) ([]*domain.Interacao, error)
	Delete(ctx context.Context, id, contaID, usuarioAtorID string, papel domain.Role) error
}
