package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type cadastroPostgres struct{ db *sql.DB }

func NewCadastroRepository(db *sql.DB) ports.CadastroRepository {
	return &cadastroPostgres{db: db}
}

func (r *cadastroPostgres) CriarContaEUsuario(ctx context.Context, conta *domain.ContaSaaS, usuario *domain.Usuario) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar cadastro: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `INSERT INTO contas_saas (tipo_conta, nome_conta, status_plano, plano, trial_vence_em)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, feature_flags, trial_vence_em, lead_estrategia, lead_token_integracao,
		billing_customer_id, billing_subscription_id, billing_status, billing_periodo_fim, criado_em, atualizado_em`,
		conta.TipoConta, conta.NomeConta, conta.StatusPlano, conta.Plano, conta.TrialVenceEm,
	).Scan(&conta.ID, &conta.FeatureFlags, &conta.TrialVenceEm, &conta.LeadEstrategia,
		&conta.LeadTokenIntegracao, &conta.BillingCustomerID, &conta.BillingSubscriptionID,
		&conta.BillingStatus, &conta.BillingPeriodoFim, &conta.CriadoEm, &conta.AtualizadoEm)
	if err != nil {
		return fmt.Errorf("criar conta: %w", err)
	}

	usuario.ContaID = conta.ID
	err = tx.QueryRowContext(ctx, `INSERT INTO usuarios (conta_id, nome_completo, email, senha_hash, google_id, papel, status, url_avatar)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, versao_sessao, criado_em, atualizado_em`,
		usuario.ContaID, usuario.NomeCompleto, usuario.Email, usuario.SenhaHash, usuario.GoogleID,
		usuario.Papel, usuario.Status, usuario.URLAvatar,
	).Scan(&usuario.ID, &usuario.VersaoSessao, &usuario.CriadoEm, &usuario.AtualizadoEm)
	if err != nil {
		return fmt.Errorf("criar usuário inicial: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar cadastro: %w", err)
	}
	return nil
}
