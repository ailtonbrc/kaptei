package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type clientePostgres struct {
	db *sql.DB
}

func NewClientePostgres(db *sql.DB) ports.ClienteRepository {
	return &clientePostgres{db: db}
}

func (r *clientePostgres) Create(ctx context.Context, cliente *domain.Cliente) error {
	tagsJSON, prefsJSON, finanJSON, utmJSON, err := serializarDadosCliente(cliente)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO clientes (
			conta_id, nome, email, telefone, status_funil, 
			origem, interesse_tipo, notas, tags, preferencias,
			cpf, data_nascimento, estado_civil, financeiro, origem_utm,
			corretor_id, temperatura, proxima_acao
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, criado_em, atualizado_em
	`
	err = r.db.QueryRowContext(ctx, query,
		cliente.ContaID,
		cliente.Nome,
		cliente.Email,
		cliente.Telefone,
		cliente.StatusFunil,
		cliente.Origem,
		cliente.InteresseTipo,
		cliente.Notas,
		tagsJSON,
		prefsJSON,
		cliente.CPF,
		cliente.DataNascimento,
		cliente.EstadoCivil,
		finanJSON,
		utmJSON,
		cliente.CorretorID,
		cliente.Temperatura,
		cliente.ProximaAcao,
	).Scan(&cliente.ID, &cliente.CriadoEm, &cliente.AtualizadoEm)

	if err != nil {
		return err
	}
	return nil
}

func (r *clientePostgres) GetByID(ctx context.Context, id, contaID string) (*domain.Cliente, error) {
	query := `
		SELECT id, conta_id, nome, email, telefone, status_funil, 
		       origem, interesse_tipo, notas, COALESCE(tags, '[]'::jsonb), COALESCE(preferencias, '{}'::jsonb),
		       cpf, data_nascimento, estado_civil, COALESCE(financeiro, '{}'::jsonb), COALESCE(origem_utm, '{}'::jsonb),
		       corretor_id, temperatura, proxima_acao, criado_em, atualizado_em
		FROM clientes
		WHERE id = $1 AND conta_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, id, contaID)

	var c domain.Cliente
	var tagsJSON, prefsJSON, finanJSON, utmJSON []byte

	err := row.Scan(
		&c.ID,
		&c.ContaID,
		&c.Nome,
		&c.Email,
		&c.Telefone,
		&c.StatusFunil,
		&c.Origem,
		&c.InteresseTipo,
		&c.Notas,
		&tagsJSON,
		&prefsJSON,
		&c.CPF,
		&c.DataNascimento,
		&c.EstadoCivil,
		&finanJSON,
		&utmJSON,
		&c.CorretorID,
		&c.Temperatura,
		&c.ProximaAcao,
		&c.CriadoEm,
		&c.AtualizadoEm,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := decodificarDadosCliente(&c, tagsJSON, prefsJSON, finanJSON, utmJSON); err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *clientePostgres) ListByContaID(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Cliente], error) {
	deslocamento := (filtro.Pagina - 1) * filtro.Limite
	condicoes := `conta_id = $1
		AND ($2::uuid IS NULL OR corretor_id = $2)
		AND ($3 = '' OR nome ILIKE '%' || $3 || '%' OR COALESCE(email, '') ILIKE '%' || $3 || '%' OR COALESCE(telefone, '') ILIKE '%' || $3 || '%')
		AND ($4 = '' OR UPPER(status_funil) = UPPER($4))`
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM clientes WHERE `+condicoes,
		contaID, filtro.UsuarioID, filtro.Busca, filtro.Status).Scan(&total); err != nil {
		return nil, fmt.Errorf("contar clientes: %w", err)
	}
	query := `
		SELECT id, conta_id, nome, email, telefone, status_funil, 
		       origem, interesse_tipo, notas, COALESCE(tags, '[]'::jsonb), COALESCE(preferencias, '{}'::jsonb),
		       cpf, data_nascimento, estado_civil, COALESCE(financeiro, '{}'::jsonb), COALESCE(origem_utm, '{}'::jsonb),
		       corretor_id, temperatura, proxima_acao, criado_em, atualizado_em
		FROM clientes
		WHERE ` + condicoes + `
		ORDER BY atualizado_em DESC, id DESC
		LIMIT $5 OFFSET $6
	`
	rows, err := r.db.QueryContext(ctx, query, contaID, filtro.UsuarioID, filtro.Busca, filtro.Status, filtro.Limite, deslocamento)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clientes []*domain.Cliente
	for rows.Next() {
		var c domain.Cliente
		var tagsJSON, prefsJSON, finanJSON, utmJSON []byte

		if err := rows.Scan(
			&c.ID,
			&c.ContaID,
			&c.Nome,
			&c.Email,
			&c.Telefone,
			&c.StatusFunil,
			&c.Origem,
			&c.InteresseTipo,
			&c.Notas,
			&tagsJSON,
			&prefsJSON,
			&c.CPF,
			&c.DataNascimento,
			&c.EstadoCivil,
			&finanJSON,
			&utmJSON,
			&c.CorretorID,
			&c.Temperatura,
			&c.ProximaAcao,
			&c.CriadoEm,
			&c.AtualizadoEm,
		); err != nil {
			return nil, err
		}

		if err := decodificarDadosCliente(&c, tagsJSON, prefsJSON, finanJSON, utmJSON); err != nil {
			return nil, err
		}

		clientes = append(clientes, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("percorrer clientes: %w", err)
	}
	if clientes == nil {
		clientes = []*domain.Cliente{}
	}
	return &domain.ListaPaginada[*domain.Cliente]{Dados: clientes, Total: total, Pagina: filtro.Pagina, Limite: filtro.Limite}, nil
}

func (r *clientePostgres) Update(ctx context.Context, cliente *domain.Cliente) error {
	tagsJSON, prefsJSON, finanJSON, utmJSON, err := serializarDadosCliente(cliente)
	if err != nil {
		return err
	}

	query := `
		UPDATE clientes
		SET nome = $1, email = $2, telefone = $3, status_funil = $4,
		    origem = $5, interesse_tipo = $6, notas = $7, tags = $8, preferencias = $9,
		    cpf = $10, data_nascimento = $11, estado_civil = $12, financeiro = $13, origem_utm = $14,
		    corretor_id = $15, temperatura = $16, proxima_acao = $17, atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $18 AND conta_id = $19
		RETURNING atualizado_em
	`
	err = r.db.QueryRowContext(ctx, query,
		cliente.Nome,
		cliente.Email,
		cliente.Telefone,
		cliente.StatusFunil,
		cliente.Origem,
		cliente.InteresseTipo,
		cliente.Notas,
		tagsJSON,
		prefsJSON,
		cliente.CPF,
		cliente.DataNascimento,
		cliente.EstadoCivil,
		finanJSON,
		utmJSON,
		cliente.CorretorID,
		cliente.Temperatura,
		cliente.ProximaAcao,
		cliente.ID,
		cliente.ContaID,
	).Scan(&cliente.AtualizadoEm)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("cliente não encontrado")
		}
		return err
	}
	return nil
}

func serializarDadosCliente(cliente *domain.Cliente) ([]byte, []byte, []byte, []byte, error) {
	serializar := func(valor any, vazio string) ([]byte, error) {
		if valor == nil {
			return []byte(vazio), nil
		}
		return json.Marshal(valor)
	}
	tags, err := serializar(cliente.Tags, "[]")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("serializar tags do cliente: %w", err)
	}
	preferencias, err := serializar(cliente.Preferencias, "{}")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("serializar preferências do cliente: %w", err)
	}
	financeiro, err := serializar(cliente.Financeiro, "{}")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("serializar dados financeiros do cliente: %w", err)
	}
	utm, err := serializar(cliente.OrigemUTM, "{}")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("serializar origem do cliente: %w", err)
	}
	return tags, preferencias, financeiro, utm, nil
}

func decodificarDadosCliente(cliente *domain.Cliente, tags, preferencias, financeiro, utm []byte) error {
	if err := json.Unmarshal(tags, &cliente.Tags); err != nil {
		return fmt.Errorf("decodificar tags do cliente: %w", err)
	}
	if string(preferencias) != "null" && string(preferencias) != "{}" {
		cliente.Preferencias = &domain.ClientePreferencias{}
		if err := json.Unmarshal(preferencias, cliente.Preferencias); err != nil {
			return fmt.Errorf("decodificar preferências do cliente: %w", err)
		}
	}
	if string(financeiro) != "null" && string(financeiro) != "{}" {
		cliente.Financeiro = &domain.ClienteFinanceiro{}
		if err := json.Unmarshal(financeiro, cliente.Financeiro); err != nil {
			return fmt.Errorf("decodificar dados financeiros do cliente: %w", err)
		}
	}
	if string(utm) != "null" && string(utm) != "{}" {
		cliente.OrigemUTM = &domain.ClienteOrigemUTM{}
		if err := json.Unmarshal(utm, cliente.OrigemUTM); err != nil {
			return fmt.Errorf("decodificar origem do cliente: %w", err)
		}
	}
	return nil
}

func (r *clientePostgres) Delete(ctx context.Context, id, contaID string) error {
	query := "DELETE FROM clientes WHERE id = $1 AND conta_id = $2"
	res, err := r.db.ExecContext(ctx, query, id, contaID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("cliente não encontrado")
	}

	return nil
}

func (r *clientePostgres) GetUltimoCorretorAtribuido(ctx context.Context, contaID string) (*string, error) {
	query := `
		SELECT corretor_id 
		FROM clientes 
		WHERE conta_id = $1 AND corretor_id IS NOT NULL 
		ORDER BY criado_em DESC 
		LIMIT 1
	`
	var corretorID string
	err := r.db.QueryRowContext(ctx, query, contaID).Scan(&corretorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &corretorID, nil
}
