package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/msdev/kaptei/internal/core/ports"
)

type billingPostgres struct{ db *sql.DB }

func NewBillingPostgres(db *sql.DB) *billingPostgres { return &billingPostgres{db: db} }

func (r *billingPostgres) ProcessarEvento(ctx context.Context, evento ports.EventoPagamento) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("iniciar processamento do evento: %w", err)
	}
	defer tx.Rollback()

	var registroID string
	err = tx.QueryRowContext(ctx, `INSERT INTO billing_eventos (provedor, evento_id, tipo, conta_id)
		VALUES ('STRIPE', $1, $2, NULLIF($3, '')::uuid)
		ON CONFLICT (provedor, evento_id) DO NOTHING RETURNING id`, evento.ID, evento.Tipo, evento.ContaID).Scan(&registroID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("registrar evento de cobrança: %w", err)
	}

	statusPlano, deveAtualizarStatus := statusPlanoPorEvento(evento)
	deveVincular := evento.Tipo == "checkout.session.completed"
	if deveAtualizarStatus || deveVincular {
		resultado, err := tx.ExecContext(ctx, `UPDATE contas_saas SET
			status_plano=COALESCE(NULLIF($1, ''), status_plano), plano=COALESCE(NULLIF($6, ''), plano), billing_status=NULLIF($2, ''),
			billing_customer_id=COALESCE(NULLIF($3, ''), billing_customer_id),
			billing_subscription_id=COALESCE(NULLIF($4, ''), billing_subscription_id),
			billing_periodo_fim=COALESCE($7, billing_periodo_fim), atualizado_em=NOW()
			WHERE ($5 <> '' AND id=$5::uuid)
			   OR ($3 <> '' AND billing_customer_id=$3)
			   OR ($4 <> '' AND billing_subscription_id=$4)`,
			statusPlano, evento.Status, evento.CustomerID, evento.SubscriptionID, evento.ContaID, evento.PlanoCodigo, evento.PeriodoFim)
		if err != nil {
			return false, fmt.Errorf("atualizar assinatura: %w", err)
		}
		linhas, err := resultado.RowsAffected()
		if err != nil || linhas != 1 {
			return false, errors.New("evento de cobrança sem conta correspondente")
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE billing_eventos SET processado_em=NOW() WHERE id=$1`, registroID); err != nil {
		return false, fmt.Errorf("concluir evento de cobrança: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("confirmar evento de cobrança: %w", err)
	}
	return true, nil
}

func statusPlanoPorEvento(evento ports.EventoPagamento) (string, bool) {
	switch evento.Tipo {
	case "invoice.paid":
		return "ATIVO", true
	case "invoice.payment_failed":
		return "INADIMPLENTE", true
	case "customer.subscription.deleted":
		return "CANCELADO", true
	case "customer.subscription.updated":
		switch strings.ToLower(evento.Status) {
		case "active", "trialing":
			return "ATIVO", true
		case "past_due", "unpaid":
			return "INADIMPLENTE", true
		case "canceled", "incomplete_expired":
			return "CANCELADO", true
		}
	}
	return "", false
}
