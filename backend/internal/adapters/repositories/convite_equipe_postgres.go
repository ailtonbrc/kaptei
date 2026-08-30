package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
)

type conviteEquipePostgres struct{ db *sql.DB }

func NewConviteEquipeRepository(db *sql.DB) *conviteEquipePostgres {
	return &conviteEquipePostgres{db: db}
}

func (r *conviteEquipePostgres) Criar(ctx context.Context, convite *domain.ConviteEquipe, limiteCorretores *int, evento *domain.EventoOutbox) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar convite de equipe: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, convite.ContaID); err != nil {
		return fmt.Errorf("bloquear vagas da equipe: %w", err)
	}
	if limiteCorretores != nil {
		var ocupadas int
		if err := tx.QueryRowContext(ctx, `SELECT
			(SELECT COUNT(*) FROM usuarios WHERE conta_id=$1 AND papel='CORRETOR_EQUIPE' AND UPPER(status)='ATIVO') +
			(SELECT COUNT(*) FROM convites_equipe WHERE conta_id=$1 AND usado_em IS NULL AND revogado_em IS NULL AND expira_em>NOW() AND LOWER(email)<>LOWER($2))`,
			convite.ContaID, convite.Email).Scan(&ocupadas); err != nil {
			return fmt.Errorf("contar vagas da equipe: %w", err)
		}
		if ocupadas >= *limiteCorretores {
			return errors.New("limite de corretores do plano atingido")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE convites_equipe SET revogado_em=NOW()
		WHERE conta_id=$1 AND LOWER(email)=LOWER($2) AND usado_em IS NULL AND revogado_em IS NULL`, convite.ContaID, convite.Email); err != nil {
		return fmt.Errorf("revogar convite anterior: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `INSERT INTO convites_equipe
		(conta_id,email,papel,token_hash,convidado_por,expira_em)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id,criado_em`,
		convite.ContaID, convite.Email, convite.Papel, convite.TokenHash, convite.ConvidadoPor, convite.ExpiraEm,
	).Scan(&convite.ID, &convite.CriadoEm); err != nil {
		return fmt.Errorf("criar convite de equipe: %w", err)
	}
	if err := inserirEventoOutbox(ctx, tx, evento); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar convite de equipe: %w", err)
	}
	return nil
}

func (r *conviteEquipePostgres) ListarPendentes(ctx context.Context, contaID string) ([]*domain.ConviteEquipe, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,conta_id,email,papel,convidado_por,expira_em,usado_em,revogado_em,criado_em
		FROM convites_equipe WHERE conta_id=$1 AND usado_em IS NULL AND revogado_em IS NULL AND expira_em>NOW()
		ORDER BY criado_em DESC`, contaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	convites := make([]*domain.ConviteEquipe, 0)
	for rows.Next() {
		convite := &domain.ConviteEquipe{}
		if err := rows.Scan(&convite.ID, &convite.ContaID, &convite.Email, &convite.Papel, &convite.ConvidadoPor,
			&convite.ExpiraEm, &convite.UsadoEm, &convite.RevogadoEm, &convite.CriadoEm); err != nil {
			return nil, err
		}
		convites = append(convites, convite)
	}
	return convites, rows.Err()
}

func (r *conviteEquipePostgres) Cancelar(ctx context.Context, conviteID, contaID string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE convites_equipe SET revogado_em=NOW()
		WHERE id=$1 AND conta_id=$2 AND usado_em IS NULL AND revogado_em IS NULL`, conviteID, contaID)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas != 1 {
		return errors.New("convite pendente não encontrado")
	}
	return nil
}

func (r *conviteEquipePostgres) Aceitar(ctx context.Context, tokenHash, nome, senhaHash string) (*domain.Usuario, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("iniciar aceite do convite: %w", err)
	}
	defer tx.Rollback()

	var convite domain.ConviteEquipe
	err = tx.QueryRowContext(ctx, `SELECT id,conta_id,email,papel,expira_em
		FROM convites_equipe WHERE token_hash=$1 AND usado_em IS NULL AND revogado_em IS NULL FOR UPDATE`, tokenHash).Scan(
		&convite.ID, &convite.ContaID, &convite.Email, &convite.Papel, &convite.ExpiraEm,
	)
	if errors.Is(err, sql.ErrNoRows) || (!convite.ExpiraEm.IsZero() && !convite.ExpiraEm.After(time.Now())) {
		return nil, errors.New("convite inválido, expirado ou já utilizado")
	}
	if err != nil {
		return nil, fmt.Errorf("carregar convite: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, convite.ContaID); err != nil {
		return nil, fmt.Errorf("bloquear vagas da equipe: %w", err)
	}
	var limite sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT p.limite_corretores FROM contas_saas c JOIN planos p ON p.codigo=c.plano WHERE c.id=$1`, convite.ContaID).Scan(&limite); err != nil {
		return nil, fmt.Errorf("carregar limite do plano: %w", err)
	}
	if limite.Valid {
		var ativos int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM usuarios WHERE conta_id=$1 AND papel='CORRETOR_EQUIPE' AND UPPER(status)='ATIVO'`, convite.ContaID).Scan(&ativos); err != nil {
			return nil, err
		}
		if int64(ativos) >= limite.Int64 {
			return nil, errors.New("limite de corretores do plano atingido")
		}
	}
	usuario := &domain.Usuario{ContaID: convite.ContaID, NomeCompleto: nome, Email: strings.ToLower(convite.Email), Papel: convite.Papel, Status: "ATIVO"}
	if err := tx.QueryRowContext(ctx, `INSERT INTO usuarios (conta_id,nome_completo,email,senha_hash,papel,status)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id,versao_sessao,criado_em,atualizado_em`,
		usuario.ContaID, usuario.NomeCompleto, usuario.Email, senhaHash, usuario.Papel, usuario.Status,
	).Scan(&usuario.ID, &usuario.VersaoSessao, &usuario.CriadoEm, &usuario.AtualizadoEm); err != nil {
		return nil, fmt.Errorf("criar usuário convidado: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE convites_equipe SET usado_em=NOW() WHERE id=$1`, convite.ID); err != nil {
		return nil, fmt.Errorf("consumir convite: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("confirmar aceite do convite: %w", err)
	}
	return usuario, nil
}
