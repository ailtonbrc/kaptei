package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type AgendamentoPostgres struct {
	db *sql.DB
}

func NewAgendamentoPostgres(db *sql.DB) ports.AgendamentoRepository {
	return &AgendamentoPostgres{db: db}
}

func (r *AgendamentoPostgres) Create(ctx context.Context, a *domain.Agendamento) error {
	query := `
		INSERT INTO agendamentos (conta_id, usuario_id, cliente_id, imovel_id, titulo, descricao, data_hora_inicio, data_hora_fim, status, tipo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		a.ContaID, a.UsuarioID, a.ClienteID, a.ImovelID, a.Titulo, a.Descricao,
		a.DataHoraInicio, a.DataHoraFim, a.Status, a.Tipo,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)

	return err
}

func (r *AgendamentoPostgres) GetByID(ctx context.Context, id, contaID string) (*domain.Agendamento, error) {
	query := `
		SELECT id, conta_id, usuario_id, cliente_id, imovel_id, titulo, descricao, data_hora_inicio, data_hora_fim, status, tipo, created_at, updated_at
		FROM agendamentos
		WHERE id = $1 AND conta_id = $2
	`
	var a domain.Agendamento
	err := r.db.QueryRowContext(ctx, query, id, contaID).Scan(
		&a.ID, &a.ContaID, &a.UsuarioID, &a.ClienteID, &a.ImovelID, &a.Titulo, &a.Descricao,
		&a.DataHoraInicio, &a.DataHoraFim, &a.Status, &a.Tipo, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AgendamentoPostgres) List(ctx context.Context, contaID string, usuarioID *string, inicio, fim time.Time) ([]*domain.Agendamento, error) {
	query := `
		SELECT a.id, a.conta_id, a.usuario_id, a.cliente_id, a.imovel_id, a.titulo, a.descricao, 
		       a.data_hora_inicio, a.data_hora_fim, a.status, a.tipo, a.created_at, a.updated_at,
			   c.nome as cliente_nome, i.titulo as imovel_titulo
		FROM agendamentos a
		LEFT JOIN clientes c ON a.cliente_id = c.id AND a.conta_id = c.conta_id
		LEFT JOIN imoveis i ON a.imovel_id = i.id AND a.conta_id = i.conta_id
		WHERE a.conta_id = $1 
		  AND a.data_hora_inicio < $3 AND a.data_hora_fim > $2
	`
	var args []interface{}
	args = append(args, contaID, inicio, fim)

	if usuarioID != nil && *usuarioID != "" {
		query += ` AND a.usuario_id = $4`
		args = append(args, *usuarioID)
	}

	query += ` ORDER BY a.data_hora_inicio ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agendamentos []*domain.Agendamento
	for rows.Next() {
		var a domain.Agendamento
		var clienteNome, imovelTitulo sql.NullString

		err := rows.Scan(
			&a.ID, &a.ContaID, &a.UsuarioID, &a.ClienteID, &a.ImovelID, &a.Titulo, &a.Descricao,
			&a.DataHoraInicio, &a.DataHoraFim, &a.Status, &a.Tipo, &a.CreatedAt, &a.UpdatedAt,
			&clienteNome, &imovelTitulo,
		)
		if err != nil {
			return nil, err
		}

		if clienteNome.Valid {
			a.ClienteNome = &clienteNome.String
		}
		if imovelTitulo.Valid {
			a.ImovelTitulo = &imovelTitulo.String
		}

		agendamentos = append(agendamentos, &a)
	}

	return agendamentos, nil
}

func (r *AgendamentoPostgres) Update(ctx context.Context, a *domain.Agendamento) error {
	query := `
		UPDATE agendamentos 
		SET titulo = $1, descricao = $2, data_hora_inicio = $3, data_hora_fim = $4, status = $5, tipo = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7 AND conta_id = $8
	`
	resultado, err := r.db.ExecContext(ctx, query,
		a.Titulo, a.Descricao, a.DataHoraInicio, a.DataHoraFim, a.Status, a.Tipo, a.ID, a.ContaID,
	)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return errors.New("agendamento não encontrado")
	}
	return nil
}

func (r *AgendamentoPostgres) Delete(ctx context.Context, id, contaID string) error {
	query := `DELETE FROM agendamentos WHERE id = $1 AND conta_id = $2`
	resultado, err := r.db.ExecContext(ctx, query, id, contaID)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return errors.New("agendamento não encontrado")
	}
	return nil
}
