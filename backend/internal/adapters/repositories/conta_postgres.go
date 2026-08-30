package repositories

import (
	"context"
	"database/sql"

	"github.com/msdev/kaptei/internal/core/domain"
)

type contaPostgres struct{ db *sql.DB }

func NewContaRepository(db *sql.DB) *contaPostgres { return &contaPostgres{db: db} }

const selecionarConta = `SELECT id, tipo_conta, nome_conta, status_plano, plano, trial_vence_em,
	feature_flags, lead_estrategia, lead_token_integracao, billing_customer_id,
	billing_subscription_id, billing_status, billing_periodo_fim, lead_token_hash, lead_token_prefixo, criado_em, atualizado_em
	FROM contas_saas`

func (r *contaPostgres) Create(ctx context.Context, conta *domain.ContaSaaS) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO contas_saas (tipo_conta, nome_conta, status_plano, plano, trial_vence_em)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, feature_flags, trial_vence_em, lead_estrategia, lead_token_integracao,
		billing_customer_id, billing_subscription_id, billing_status, billing_periodo_fim,
		lead_token_hash, lead_token_prefixo, criado_em, atualizado_em`,
		conta.TipoConta, conta.NomeConta, conta.StatusPlano, conta.Plano, conta.TrialVenceEm,
	).Scan(&conta.ID, &conta.FeatureFlags, &conta.TrialVenceEm, &conta.LeadEstrategia,
		&conta.LeadTokenIntegracao, &conta.BillingCustomerID, &conta.BillingSubscriptionID,
		&conta.BillingStatus, &conta.BillingPeriodoFim, &conta.LeadTokenHash, &conta.LeadTokenPrefixo,
		&conta.CriadoEm, &conta.AtualizadoEm)
}

func (r *contaPostgres) GetByID(ctx context.Context, id string) (*domain.ContaSaaS, error) {
	return r.buscar(ctx, selecionarConta+` WHERE id = $1`, id)
}

func (r *contaPostgres) GetByLeadToken(ctx context.Context, token string) (*domain.ContaSaaS, error) {
	return r.buscar(ctx, selecionarConta+` WHERE lead_token_integracao = $1`, token)
}

func (r *contaPostgres) GetByLeadTokenHash(ctx context.Context, tokenHash string) (*domain.ContaSaaS, error) {
	return r.buscar(ctx, selecionarConta+` WHERE lead_token_hash = $1`, tokenHash)
}

func (r *contaPostgres) buscar(ctx context.Context, query, valor string) (*domain.ContaSaaS, error) {
	conta := &domain.ContaSaaS{}
	err := r.db.QueryRowContext(ctx, query, valor).Scan(
		&conta.ID, &conta.TipoConta, &conta.NomeConta, &conta.StatusPlano, &conta.Plano,
		&conta.TrialVenceEm, &conta.FeatureFlags, &conta.LeadEstrategia, &conta.LeadTokenIntegracao,
		&conta.BillingCustomerID, &conta.BillingSubscriptionID, &conta.BillingStatus,
		&conta.BillingPeriodoFim, &conta.LeadTokenHash, &conta.LeadTokenPrefixo,
		&conta.CriadoEm, &conta.AtualizadoEm,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return conta, nil
}

func (r *contaPostgres) AtualizarEstrategiaLeads(ctx context.Context, contaID, estrategia string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE contas_saas SET lead_estrategia=$1,atualizado_em=NOW() WHERE id=$2`, estrategia, contaID)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *contaPostgres) RotacionarTokenLeads(ctx context.Context, contaID, tokenHash, prefixo string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE contas_saas SET
		lead_token_integracao=NULL,lead_token_hash=$1,lead_token_prefixo=$2,atualizado_em=NOW() WHERE id=$3`, tokenHash, prefixo, contaID)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *contaPostgres) Update(ctx context.Context, conta *domain.ContaSaaS) error {
	_, err := r.db.ExecContext(ctx, `UPDATE contas_saas SET
		tipo_conta=$1, nome_conta=$2, status_plano=$3, plano=$4, trial_vence_em=$5,
		feature_flags=$6, lead_estrategia=$7, billing_customer_id=$8,
		billing_subscription_id=$9, billing_status=$10, billing_periodo_fim=$11, atualizado_em=NOW()
		WHERE id=$12`, conta.TipoConta, conta.NomeConta, conta.StatusPlano, conta.Plano,
		conta.TrialVenceEm, conta.FeatureFlags, conta.LeadEstrategia, conta.BillingCustomerID,
		conta.BillingSubscriptionID, conta.BillingStatus, conta.BillingPeriodoFim, conta.ID)
	return err
}
