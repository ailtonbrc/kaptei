package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type configuracaoPostgres struct {
	db *sql.DB
}

func NewConfiguracaoRepository(db *sql.DB) ports.ConfiguracaoRepository {
	return &configuracaoPostgres{db: db}
}

func (r *configuracaoPostgres) Get(ctx context.Context, chave string) (*domain.ConfiguracaoSistema, error) {
	query := `
		SELECT chave, valor, descricao, atualizado_em
		FROM configuracoes_sistema
		WHERE chave = $1
	`

	var config domain.ConfiguracaoSistema
	err := r.db.QueryRowContext(ctx, query, chave).Scan(
		&config.Chave, &config.Valor, &config.Descricao, &config.AtualizadoEm,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Não encontrou a configuração
		}
		return nil, err
	}

	return &config, nil
}

func (r *configuracaoPostgres) Set(ctx context.Context, config *domain.ConfiguracaoSistema) error {
	query := `
		INSERT INTO configuracoes_sistema (chave, valor, descricao, atualizado_em)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chave) DO UPDATE
		SET valor = EXCLUDED.valor,
		    descricao = EXCLUDED.descricao,
		    atualizado_em = $4
	`

	_, err := r.db.ExecContext(ctx, query,
		config.Chave,
		config.Valor,
		config.Descricao,
		time.Now(),
	)

	return err
}
