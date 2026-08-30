package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/msdev/kaptei/internal/core/domain"
)

type sessaoPostgres struct{ db *sql.DB }

func NewSessaoRepository(db *sql.DB) *sessaoPostgres {
	return &sessaoPostgres{db: db}
}

func (r *sessaoPostgres) Criar(ctx context.Context, sessao *domain.SessaoUsuario) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar sessão: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, sessao.UsuarioID); err != nil {
		return fmt.Errorf("bloquear sessões do usuário: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessoes_usuario WHERE usuario_id=$1 AND (expira_em<=NOW() OR revogada_em IS NOT NULL)`, sessao.UsuarioID); err != nil {
		return fmt.Errorf("limpar sessões expiradas: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessoes_usuario SET revogada_em=NOW() WHERE id IN (
		SELECT id FROM sessoes_usuario WHERE usuario_id=$1 AND revogada_em IS NULL
		ORDER BY criado_em DESC OFFSET 9
	)`, sessao.UsuarioID); err != nil {
		return fmt.Errorf("limitar sessões simultâneas: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `INSERT INTO sessoes_usuario (usuario_id,conta_id,expira_em)
		VALUES ($1,$2,$3) RETURNING id,criado_em`, sessao.UsuarioID, sessao.ContaID, sessao.ExpiraEm).Scan(&sessao.ID, &sessao.CriadoEm); err != nil {
		return fmt.Errorf("criar sessão: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar sessão: %w", err)
	}
	return nil
}

func (r *sessaoPostgres) EstaAtiva(ctx context.Context, sessaoID, usuarioID string) (bool, error) {
	var ativa bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM sessoes_usuario WHERE id=$1 AND usuario_id=$2 AND revogada_em IS NULL AND expira_em>NOW()
	)`, sessaoID, usuarioID).Scan(&ativa)
	return ativa, err
}

func (r *sessaoPostgres) Revogar(ctx context.Context, sessaoID, usuarioID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE sessoes_usuario SET revogada_em=COALESCE(revogada_em,NOW()) WHERE id=$1 AND usuario_id=$2`, sessaoID, usuarioID)
	return err
}
