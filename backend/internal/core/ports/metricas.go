package ports

import (
	"context"
	"time"
)

type MetricasAplicacao interface {
	RegistrarHTTP(ctx context.Context, metodo, rota string, status int, duracao time.Duration)
	RegistrarItemFila(ctx context.Context, fila, resultado string, duracao time.Duration)
}

type MetricasNulas struct{}

func (MetricasNulas) RegistrarHTTP(context.Context, string, string, int, time.Duration) {}
func (MetricasNulas) RegistrarItemFila(context.Context, string, string, time.Duration)  {}
