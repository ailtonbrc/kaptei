package repositories

import (
	"context"
	"database/sql"

	"github.com/msdev/kaptei/internal/core/domain"
)

type PlanoRepository struct{ db *sql.DB }

func NewPlanoRepository(db *sql.DB) *PlanoRepository { return &PlanoRepository{db: db} }

const selecionarPlano = `SELECT id, codigo, tipo, nome, subtitle, preco, cor, recomendado, features, missing, ativo, criado_em, atualizado_em, gateway_price_id, limite_corretores FROM planos`

func (r *PlanoRepository) ListarAtivos(ctx context.Context) ([]domain.Plano, error) {
	rows, err := r.db.QueryContext(ctx, selecionarPlano+` WHERE ativo = TRUE ORDER BY preco ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	planos := make([]domain.Plano, 0)
	for rows.Next() {
		var plano domain.Plano
		if err := scanPlano(rows, &plano); err != nil {
			return nil, err
		}
		planos = append(planos, plano)
	}
	return planos, rows.Err()
}

func (r *PlanoRepository) GetByCodigo(ctx context.Context, codigo string) (*domain.Plano, error) {
	var plano domain.Plano
	err := scanPlano(r.db.QueryRowContext(ctx, selecionarPlano+` WHERE codigo = $1`, codigo), &plano)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &plano, nil
}

func (r *PlanoRepository) AtualizarGatewayPriceID(ctx context.Context, codigo, priceID string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE planos SET gateway_price_id=NULLIF($1, ''), atualizado_em=NOW() WHERE codigo=$2`, priceID, codigo)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas != 1 {
		return sql.ErrNoRows
	}
	return nil
}

type planoScanner interface{ Scan(dest ...any) error }

func scanPlano(origem planoScanner, plano *domain.Plano) error {
	return origem.Scan(&plano.ID, &plano.Codigo, &plano.Tipo, &plano.Nome, &plano.Subtitle, &plano.Preco,
		&plano.Cor, &plano.Recomendado, &plano.Features, &plano.Missing, &plano.Ativo,
		&plano.CriadoEm, &plano.AtualizadoEm, &plano.GatewayPriceID, &plano.LimiteCorretores)
}
