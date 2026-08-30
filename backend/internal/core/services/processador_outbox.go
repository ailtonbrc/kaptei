package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type ConfiguracaoProcessadorOutbox struct {
	TrabalhadorID   string
	Intervalo       time.Duration
	TamanhoLote     int
	DuracaoBloqueio time.Duration
	BackoffInicial  time.Duration
	BackoffMaximo   time.Duration
	Metricas        ports.MetricasAplicacao
}

type processadorOutbox struct {
	repositorio ports.OutboxRepository
	tratadores  map[string]ports.TratadorEventoOutbox
	config      ConfiguracaoProcessadorOutbox
}

func NewProcessadorOutbox(
	repositorio ports.OutboxRepository,
	tratadores []ports.TratadorEventoOutbox,
	config ConfiguracaoProcessadorOutbox,
) *processadorOutbox {
	porTipo := make(map[string]ports.TratadorEventoOutbox, len(tratadores))
	for _, tratador := range tratadores {
		if tratador != nil && tratador.Tipo() != "" {
			porTipo[tratador.Tipo()] = tratador
		}
	}
	return &processadorOutbox{repositorio: repositorio, tratadores: porTipo, config: config}
}

func (p *processadorOutbox) Executar(ctx context.Context) {
	executarProcessadorFila(ctx, "outbox", p.config.Intervalo, p.processarLote)
}

func (p *processadorOutbox) processarLote(ctx context.Context) {
	processarLoteFila(ctx, operacoesFila[*domain.EventoOutbox]{
		Nome: "outbox", Config: p.config, Reservar: p.repositorio.Reservar,
		Metadados: func(evento *domain.EventoOutbox) metadadosItemFila {
			return metadadosItemFila{ID: evento.ID, Tentativa: evento.Tentativas, MaximoTentativas: evento.MaximoTentativas}
		},
		Processar: p.processarEvento, Concluir: p.repositorio.Concluir, Falhar: p.repositorio.Falhar,
		Definitivo: func(err error) bool { return errors.Is(err, errEventoOutboxInvalido) || erroNaoRetentavel(err) },
	})
}

var errEventoOutboxInvalido = errors.New("evento de outbox inválido")

func (p *processadorOutbox) processarEvento(ctx context.Context, evento *domain.EventoOutbox) error {
	tratador := p.tratadores[evento.Tipo]
	if tratador == nil {
		return fmt.Errorf("%w: tipo %q não suportado", errEventoOutboxInvalido, evento.Tipo)
	}
	return tratador.Processar(ctx, evento)
}

func (p *processadorOutbox) calcularBackoff(tentativa int) time.Duration {
	return calcularBackoffFila(tentativa, p.config.BackoffInicial, p.config.BackoffMaximo)
}
