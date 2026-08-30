package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type leadPostgres struct {
	db *sql.DB
}

type executorLead interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

const inserirLeadSQL = `INSERT INTO leads (
	conta_id, usuario_id, imovel_id, cliente_id, nome, email, telefone, origem, mensagem, status,
	pagina_origem, utm_source, utm_medium, utm_campaign, consentimento_lgpd,
	consentimento_lgpd_em, consentimento_lgpd_versao, criado_em, atualizado_em, chave_idempotencia
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,$18)
ON CONFLICT (conta_id, chave_idempotencia) WHERE chave_idempotencia IS NOT NULL DO NOTHING
RETURNING id, criado_em, atualizado_em`

func NewLeadPostgres(db *sql.DB) ports.LeadRepository {
	return &leadPostgres{db: db}
}

func (r *leadPostgres) Create(ctx context.Context, lead *domain.Lead) error {
	return criarLead(ctx, r.db, lead)
}

func criarLead(ctx context.Context, executor executorLead, lead *domain.Lead) error {
	err := executor.QueryRowContext(
		ctx, inserirLeadSQL,
		lead.ContaID, lead.UsuarioID, lead.ImovelID, lead.ClienteID, lead.Nome, lead.Email, lead.Telefone, lead.Origem, lead.Mensagem, lead.Status,
		lead.PaginaOrigem, lead.UTMSource, lead.UTMMedium, lead.UTMCampaign, lead.ConsentimentoLGPD,
		lead.ConsentimentoEm, lead.ConsentimentoVersao,
		lead.ChaveIdempotencia,
	).Scan(&lead.ID, &lead.CriadoEm, &lead.AtualizadoEm)
	if errors.Is(err, sql.ErrNoRows) && lead.ChaveIdempotencia != nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("criar lead: %w", err)
	}
	return nil
}

func (r *leadPostgres) CreateDistribuido(ctx context.Context, lead *domain.Lead) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar distribuição do lead: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `INSERT INTO lead_distribuicao_estado (conta_id) VALUES ($1) ON CONFLICT DO NOTHING`, lead.ContaID); err != nil {
		return fmt.Errorf("preparar estado da distribuição: %w", err)
	}
	var ultimo sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT ultimo_usuario_id FROM lead_distribuicao_estado WHERE conta_id=$1 FOR UPDATE`, lead.ContaID).Scan(&ultimo); err != nil {
		return fmt.Errorf("bloquear estado da distribuição: %w", err)
	}
	if lead.ChaveIdempotencia != nil {
		var existe bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM leads WHERE conta_id=$1 AND chave_idempotencia=$2)`, lead.ContaID, lead.ChaveIdempotencia).Scan(&existe); err != nil {
			return fmt.Errorf("verificar repetição do lead: %w", err)
		}
		if existe {
			return tx.Commit()
		}
	}
	var proximoID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM usuarios
		WHERE conta_id=$1 AND UPPER(status)='ATIVO' AND papel IN ('CORRETOR_EQUIPE','GESTOR')
		ORDER BY CASE WHEN papel='CORRETOR_EQUIPE' THEN 0 ELSE 1 END,
			CASE WHEN $2::uuid IS NULL OR id > $2::uuid THEN 0 ELSE 1 END, id
		LIMIT 1`, lead.ContaID, ultimo).Scan(&proximoID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("selecionar próximo corretor: %w", err)
	}
	if err == nil {
		lead.UsuarioID = &proximoID
		if _, err := tx.ExecContext(ctx, `UPDATE lead_distribuicao_estado SET ultimo_usuario_id=$1, atualizado_em=NOW() WHERE conta_id=$2`, proximoID, lead.ContaID); err != nil {
			return fmt.Errorf("avançar estado da distribuição: %w", err)
		}
	} else {
		lead.UsuarioID = nil
	}
	if err := criarLead(ctx, tx, lead); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar distribuição do lead: %w", err)
	}
	return nil
}

func (r *leadPostgres) GetByID(ctx context.Context, id, contaID string) (*domain.Lead, error) {
	query := `
		SELECT 
			id, conta_id, usuario_id, imovel_id, cliente_id, nome, email, telefone, origem, mensagem, status, motivo_descarte,
			pagina_origem, utm_source, utm_medium, utm_campaign, consentimento_lgpd,
			consentimento_lgpd_em, consentimento_lgpd_versao, criado_em, atualizado_em
		FROM leads
		WHERE id = $1 AND conta_id = $2
	`
	lead := &domain.Lead{}
	err := r.db.QueryRowContext(ctx, query, id, contaID).Scan(
		&lead.ID, &lead.ContaID, &lead.UsuarioID, &lead.ImovelID, &lead.ClienteID, &lead.Nome, &lead.Email, &lead.Telefone, &lead.Origem, &lead.Mensagem, &lead.Status, &lead.MotivoDescarte,
		&lead.PaginaOrigem, &lead.UTMSource, &lead.UTMMedium, &lead.UTMCampaign, &lead.ConsentimentoLGPD,
		&lead.ConsentimentoEm, &lead.ConsentimentoVersao, &lead.CriadoEm, &lead.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar lead por id: %v", err)
	}
	return lead, nil
}

func (r *leadPostgres) ListByContaID(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Lead], error) {
	deslocamento := (filtro.Pagina - 1) * filtro.Limite
	condicoes := `conta_id = $1
		AND ($2::uuid IS NULL OR usuario_id = $2)
		AND ($3 = '' OR nome ILIKE '%' || $3 || '%' OR COALESCE(email, '') ILIKE '%' || $3 || '%' OR COALESCE(telefone, '') ILIKE '%' || $3 || '%')
		AND ($4 = '' OR UPPER(status) = UPPER($4))`
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM leads WHERE `+condicoes,
		contaID, filtro.UsuarioID, filtro.Busca, filtro.Status).Scan(&total); err != nil {
		return nil, fmt.Errorf("contar leads: %w", err)
	}
	query := `
		SELECT 
			id, conta_id, usuario_id, imovel_id, cliente_id, nome, email, telefone, origem, mensagem, status, motivo_descarte,
			pagina_origem, utm_source, utm_medium, utm_campaign, consentimento_lgpd,
			consentimento_lgpd_em, consentimento_lgpd_versao, criado_em, atualizado_em
		FROM leads
		WHERE ` + condicoes + `
		ORDER BY criado_em DESC, id DESC
		LIMIT $5 OFFSET $6
	`
	rows, err := r.db.QueryContext(ctx, query, contaID, filtro.UsuarioID, filtro.Busca, filtro.Status, filtro.Limite, deslocamento)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar leads: %v", err)
	}
	defer rows.Close()

	var leads []*domain.Lead
	for rows.Next() {
		lead := &domain.Lead{}
		err := rows.Scan(
			&lead.ID, &lead.ContaID, &lead.UsuarioID, &lead.ImovelID, &lead.ClienteID, &lead.Nome, &lead.Email, &lead.Telefone, &lead.Origem, &lead.Mensagem, &lead.Status, &lead.MotivoDescarte,
			&lead.PaginaOrigem, &lead.UTMSource, &lead.UTMMedium, &lead.UTMCampaign, &lead.ConsentimentoLGPD,
			&lead.ConsentimentoEm, &lead.ConsentimentoVersao, &lead.CriadoEm, &lead.AtualizadoEm,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao scanear lead: %v", err)
		}
		leads = append(leads, lead)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("percorrer leads: %w", err)
	}
	if leads == nil {
		leads = []*domain.Lead{}
	}
	return &domain.ListaPaginada[*domain.Lead]{Dados: leads, Total: total, Pagina: filtro.Pagina, Limite: filtro.Limite}, nil
}

func (r *leadPostgres) Update(ctx context.Context, lead *domain.Lead) error {
	query := `
		UPDATE leads SET
			usuario_id = $1, imovel_id = $2, cliente_id = $3, nome = $4, email = $5, telefone = $6, origem = $7, mensagem = $8, status = $9, motivo_descarte = $10, atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $11 AND conta_id = $12
	`
	res, err := r.db.ExecContext(
		ctx, query,
		lead.UsuarioID, lead.ImovelID, lead.ClienteID, lead.Nome, lead.Email, lead.Telefone, lead.Origem, lead.Mensagem, lead.Status, lead.MotivoDescarte, lead.ID, lead.ContaID,
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar lead: %v", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("lead não encontrado ou sem permissão para atualizar")
	}
	return nil
}

func (r *leadPostgres) Delete(ctx context.Context, id, contaID string) error {
	query := `DELETE FROM leads WHERE id = $1 AND conta_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, contaID)
	if err != nil {
		return fmt.Errorf("erro ao deletar lead: %v", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("lead não encontrado ou sem permissão para deletar")
	}
	return nil
}

func (r *leadPostgres) Qualificar(ctx context.Context, id, contaID string) error {
	transacao, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar qualificação do lead: %w", err)
	}
	defer transacao.Rollback()

	var lead domain.Lead
	err = transacao.QueryRowContext(ctx, `
		SELECT id, conta_id, usuario_id, nome, email, telefone, origem, mensagem, status, cliente_id
		FROM leads
		WHERE id = $1 AND conta_id = $2
		FOR UPDATE
	`, id, contaID).Scan(
		&lead.ID, &lead.ContaID, &lead.UsuarioID, &lead.Nome, &lead.Email,
		&lead.Telefone, &lead.Origem, &lead.Mensagem, &lead.Status, &lead.ClienteID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("lead não encontrado")
	}
	if err != nil {
		return fmt.Errorf("carregar lead para qualificação: %w", err)
	}
	if lead.Status == domain.LeadStatusQualificado && lead.ClienteID != nil {
		return nil
	}

	origem := lead.Origem
	if origem == nil {
		valor := "OUTROS"
		origem = &valor
	}

	var clienteID string
	err = transacao.QueryRowContext(ctx, `
		INSERT INTO clientes (
			conta_id, nome, email, telefone, status_funil, origem, notas, corretor_id
		) VALUES ($1, $2, $3, $4, 'NOVO', $5, $6, $7)
		RETURNING id
	`, lead.ContaID, lead.Nome, lead.Email, lead.Telefone, origem, lead.Mensagem, lead.UsuarioID).Scan(&clienteID)
	if err != nil {
		return fmt.Errorf("criar cliente a partir do lead: %w", err)
	}

	resultado, err := transacao.ExecContext(ctx, `
		UPDATE leads
		SET status = $1, cliente_id = $2, atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $3 AND conta_id = $4
	`, domain.LeadStatusQualificado, clienteID, lead.ID, lead.ContaID)
	if err != nil {
		return fmt.Errorf("concluir qualificação do lead: %w", err)
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return errors.New("não foi possível concluir a qualificação do lead")
	}

	if err := transacao.Commit(); err != nil {
		return fmt.Errorf("confirmar qualificação do lead: %w", err)
	}
	return nil
}
