package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type recuperacaoSenhaPostgres struct {
	db *sql.DB
}

func NewRecuperacaoSenhaRepository(db *sql.DB) ports.RecuperacaoSenhaRepository {
	return &recuperacaoSenhaPostgres{db: db}
}

func (r *recuperacaoSenhaPostgres) CreateToken(ctx context.Context, token *domain.RecuperacaoSenhaToken, evento *domain.EventoOutbox) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar recuperação de senha: %w", err)
	}
	defer tx.Rollback()
	query := `
		INSERT INTO recuperacao_senha_tokens (usuario_id, token, expira_em, usado, criado_em)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	if err := tx.QueryRowContext(
		ctx,
		query,
		token.UsuarioID,
		token.Token,
		token.ExpiraEm,
		token.Usado,
		token.CriadoEm,
	).Scan(&token.ID); err != nil {
		return fmt.Errorf("criar token de recuperação: %w", err)
	}
	if err := inserirEventoOutbox(ctx, tx, evento); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar recuperação de senha: %w", err)
	}
	return nil
}

func (r *recuperacaoSenhaPostgres) GetToken(ctx context.Context, token string) (*domain.RecuperacaoSenhaToken, error) {
	query := `
		SELECT id, usuario_id, token, expira_em, usado, criado_em
		FROM recuperacao_senha_tokens
		WHERE token = $1
	`
	var t domain.RecuperacaoSenhaToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&t.ID, &t.UsuarioID, &t.Token, &t.ExpiraEm, &t.Usado, &t.CriadoEm,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Token não encontrado
		}
		return nil, err
	}
	return &t, nil
}

func (r *recuperacaoSenhaPostgres) ConsumirEAtualizarSenha(ctx context.Context, tokenID, usuarioID, senhaHash string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	resultado, err := tx.ExecContext(ctx, `UPDATE recuperacao_senha_tokens SET usado=TRUE
		WHERE id=$1 AND usuario_id=$2 AND usado=FALSE AND expira_em > NOW()`, tokenID, usuarioID)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return errors.New("token inválido, expirado ou já utilizado")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE usuarios SET senha_hash=$1, versao_sessao=versao_sessao+1, atualizado_em=NOW() WHERE id=$2`, senhaHash, usuarioID); err != nil {
		return err
	}
	return tx.Commit()
}
