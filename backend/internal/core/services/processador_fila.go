package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/msdev/kaptei/internal/core/ports"
)

type metadadosItemFila struct {
	ID               string
	Tentativa        int
	MaximoTentativas int
}

type operacoesFila[T any] struct {
	Nome       string
	Config     ConfiguracaoProcessadorOutbox
	Reservar   func(context.Context, string, int, time.Duration) ([]T, error)
	Metadados  func(T) metadadosItemFila
	Processar  func(context.Context, T) error
	Concluir   func(context.Context, string, string) error
	Falhar     func(context.Context, string, string, string, time.Time, bool) error
	Definitivo func(error) bool
}

func executarProcessadorFila(ctx context.Context, nome string, intervalo time.Duration, processar func(context.Context)) {
	processar(ctx)
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("processador encerrado", "fila", nome)
			return
		case <-ticker.C:
			processar(ctx)
		}
	}
}

func processarLoteFila[T any](ctx context.Context, operacoes operacoesFila[T]) {
	metricas := operacoes.Config.Metricas
	if metricas == nil {
		metricas = ports.MetricasNulas{}
	}
	inicioReserva := time.Now()
	itens, err := operacoes.Reservar(ctx, operacoes.Config.TrabalhadorID, operacoes.Config.TamanhoLote, operacoes.Config.DuracaoBloqueio)
	if err != nil {
		metricas.RegistrarItemFila(ctx, operacoes.Nome, "reserva_falhou", time.Since(inicioReserva))
		if !errors.Is(err, context.Canceled) {
			slog.Error("falha ao reservar fila", "fila", operacoes.Nome, "erro", err)
		}
		return
	}
	for _, item := range itens {
		if ctx.Err() != nil {
			return
		}
		metadados := operacoes.Metadados(item)
		inicioItem := time.Now()
		if err := operacoes.Processar(ctx, item); err != nil {
			definitivo := metadados.Tentativa >= metadados.MaximoTentativas
			if operacoes.Definitivo != nil && operacoes.Definitivo(err) {
				definitivo = true
			}
			proxima := time.Now().UTC().Add(calcularBackoffFila(metadados.Tentativa, operacoes.Config.BackoffInicial, operacoes.Config.BackoffMaximo))
			if registro := operacoes.Falhar(ctx, metadados.ID, operacoes.Config.TrabalhadorID, err.Error(), proxima, definitivo); registro != nil {
				slog.Error("falha ao registrar erro da fila", "fila", operacoes.Nome, "item_id", metadados.ID, "erro", registro)
			}
			slog.Warn("item de fila não processado", "fila", operacoes.Nome, "item_id", metadados.ID, "tentativa", metadados.Tentativa, "definitivo", definitivo)
			resultado := "retentativa"
			if definitivo {
				resultado = "falha_definitiva"
			}
			metricas.RegistrarItemFila(ctx, operacoes.Nome, resultado, time.Since(inicioItem))
			continue
		}
		if err := operacoes.Concluir(ctx, metadados.ID, operacoes.Config.TrabalhadorID); err != nil {
			slog.Error("falha ao concluir item da fila", "fila", operacoes.Nome, "item_id", metadados.ID, "erro", err)
			metricas.RegistrarItemFila(ctx, operacoes.Nome, "conclusao_falhou", time.Since(inicioItem))
			continue
		}
		metricas.RegistrarItemFila(ctx, operacoes.Nome, "concluido", time.Since(inicioItem))
	}
}

func calcularBackoffFila(tentativa int, inicial, maximo time.Duration) time.Duration {
	if tentativa <= 1 {
		return inicial
	}
	atraso := inicial
	for indice := 1; indice < tentativa; indice++ {
		if atraso >= maximo/2 {
			return maximo
		}
		atraso *= 2
	}
	if atraso > maximo {
		return maximo
	}
	return atraso
}
