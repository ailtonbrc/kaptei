package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type BootstrapPostgres struct{ db *sql.DB }

func NewBootstrapPostgres(db *sql.DB) *BootstrapPostgres {
	return &BootstrapPostgres{db: db}
}

func (r *BootstrapPostgres) CriarPrimeiroSuperAdmin(ctx context.Context, nome, email, senhaHash string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar bootstrap: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(127371)`); err != nil {
		return fmt.Errorf("bloquear bootstrap: %w", err)
	}
	var existe bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM usuarios WHERE papel='SUPER_ADMIN')`).Scan(&existe); err != nil {
		return fmt.Errorf("verificar superadministrador: %w", err)
	}
	if existe {
		return errors.New("já existe um superadministrador; o bootstrap foi recusado")
	}
	var contaID string
	if err := tx.QueryRowContext(ctx, `INSERT INTO contas_saas
		(tipo_conta,nome_conta,status_plano,plano,lead_estrategia)
		VALUES ('CORRETOR_SOLO','Administração Kaptei','ATIVO','GRATUITO','CAIXA_ENTRADA') RETURNING id`).Scan(&contaID); err != nil {
		return fmt.Errorf("criar conta administrativa: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO usuarios
		(conta_id,nome_completo,email,senha_hash,papel,status)
		VALUES ($1,$2,$3,$4,'SUPER_ADMIN','ATIVO')`, contaID, nome, email, senhaHash); err != nil {
		return fmt.Errorf("criar superadministrador: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar bootstrap: %w", err)
	}
	return nil
}
