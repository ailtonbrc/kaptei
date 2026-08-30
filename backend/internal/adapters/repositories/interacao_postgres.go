package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type interacaoPostgres struct {
	db *sql.DB
}

func NewInteracaoPostgres(db *sql.DB) ports.InteracaoRepository {
	return &interacaoPostgres{db: db}
}

func (r *interacaoPostgres) Create(ctx context.Context, interacao *domain.Interacao) error {
	query := `
		INSERT INTO interacoes (conta_id, cliente_id, corretor_id, tipo, descricao, data_hora)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, criado_em
	`
	err := r.db.QueryRowContext(ctx, query,
		interacao.ContaID,
		interacao.ClienteID,
		interacao.CorretorID,
		interacao.Tipo,
		interacao.Descricao,
		interacao.DataHora,
	).Scan(&interacao.ID, &interacao.CriadoEm)

	return err
}

func (r *interacaoPostgres) GetByID(ctx context.Context, id, contaID string) (*domain.Interacao, error) {
	var interacao domain.Interacao
	err := r.db.QueryRowContext(ctx, `SELECT id, conta_id, cliente_id, corretor_id, tipo, descricao, data_hora, criado_em
		FROM interacoes WHERE id=$1 AND conta_id=$2`, id, contaID).Scan(
		&interacao.ID, &interacao.ContaID, &interacao.ClienteID, &interacao.CorretorID,
		&interacao.Tipo, &interacao.Descricao, &interacao.DataHora, &interacao.CriadoEm,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &interacao, nil
}

func (r *interacaoPostgres) ListByClienteID(ctx context.Context, clienteID, contaID string) ([]*domain.Interacao, error) {
	query := `
		SELECT id, conta_id, cliente_id, corretor_id, tipo, descricao, data_hora, criado_em
		FROM interacoes
		WHERE cliente_id = $1 AND conta_id = $2
		ORDER BY data_hora DESC, criado_em DESC
	`
	rows, err := r.db.QueryContext(ctx, query, clienteID, contaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interacoes []*domain.Interacao
	for rows.Next() {
		var i domain.Interacao
		if err := rows.Scan(
			&i.ID,
			&i.ContaID,
			&i.ClienteID,
			&i.CorretorID,
			&i.Tipo,
			&i.Descricao,
			&i.DataHora,
			&i.CriadoEm,
		); err != nil {
			return nil, err
		}
		interacoes = append(interacoes, &i)
	}
	return interacoes, nil
}

func (r *interacaoPostgres) Delete(ctx context.Context, id, contaID string) error {
	query := "DELETE FROM interacoes WHERE id = $1 AND conta_id = $2"
	res, err := r.db.ExecContext(ctx, query, id, contaID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("interação não encontrada")
	}

	return nil
}
