package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/msdev/kaptei/internal/core/domain"
)

type usuarioPostgres struct {
	db *sql.DB
}

func NewUsuarioRepository(db *sql.DB) *usuarioPostgres {
	return &usuarioPostgres{db: db}
}

func (r *usuarioPostgres) Create(ctx context.Context, u *domain.Usuario) error {
	query := `
		INSERT INTO usuarios (conta_id, nome_completo, email, senha_hash, google_id, papel, status, url_avatar)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, versao_sessao, criado_em, atualizado_em`

	err := r.db.QueryRowContext(ctx, query,
		u.ContaID, u.NomeCompleto, u.Email, u.SenhaHash, u.GoogleID, u.Papel, u.Status, u.URLAvatar,
	).Scan(&u.ID, &u.VersaoSessao, &u.CriadoEm, &u.AtualizadoEm)

	return err
}

func (r *usuarioPostgres) GetByEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	return r.getByField(ctx, "email", email)
}

func (r *usuarioPostgres) GetByGoogleID(ctx context.Context, googleID string) (*domain.Usuario, error) {
	return r.getByField(ctx, "google_id", googleID)
}

func (r *usuarioPostgres) GetByID(ctx context.Context, id string) (*domain.Usuario, error) {
	return r.getByField(ctx, "id", id)
}

func (r *usuarioPostgres) getByField(ctx context.Context, field, value string) (*domain.Usuario, error) {
	var u domain.Usuario
	query := `
		SELECT id, conta_id, nome_completo, email, senha_hash, google_id, papel, status, url_avatar, versao_sessao,
		cpf, rg, rg_estado, rg_orgao_expedidor, nacionalidade, estado_civil, creci, creci_estado, cep, logradouro, numero, complemento, bairro, cidade, estado, telefone, numero_whatsapp,
		criado_em, atualizado_em
		FROM usuarios WHERE ` + field + ` = $1`

	err := r.db.QueryRowContext(ctx, query, value).Scan(
		&u.ID, &u.ContaID, &u.NomeCompleto, &u.Email, &u.SenhaHash, &u.GoogleID, &u.Papel, &u.Status, &u.URLAvatar, &u.VersaoSessao,
		&u.CPF, &u.RG, &u.RGEstado, &u.RGOrgaoExpedidor, &u.Nacionalidade, &u.EstadoCivil, &u.Creci, &u.CreciEstado, &u.CEP, &u.Logradouro, &u.Numero, &u.Complemento, &u.Bairro, &u.Cidade, &u.Estado, &u.Telefone, &u.NumeroWhatsapp,
		&u.CriadoEm, &u.AtualizadoEm,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not necessarily an error in Go, return nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *usuarioPostgres) ListByContaID(ctx context.Context, contaID string) ([]*domain.Usuario, error) {
	query := `
		SELECT id, conta_id, nome_completo, email, senha_hash, google_id, papel, status, url_avatar, versao_sessao,
		cpf, rg, rg_estado, rg_orgao_expedidor, nacionalidade, estado_civil, creci, creci_estado, cep, logradouro, numero, complemento, bairro, cidade, estado, telefone, numero_whatsapp,
		criado_em, atualizado_em
		FROM usuarios WHERE conta_id = $1 ORDER BY criado_em ASC`

	rows, err := r.db.QueryContext(ctx, query, contaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []*domain.Usuario
	for rows.Next() {
		var u domain.Usuario
		err := rows.Scan(
			&u.ID, &u.ContaID, &u.NomeCompleto, &u.Email, &u.SenhaHash, &u.GoogleID, &u.Papel, &u.Status, &u.URLAvatar, &u.VersaoSessao,
			&u.CPF, &u.RG, &u.RGEstado, &u.RGOrgaoExpedidor, &u.Nacionalidade, &u.EstadoCivil, &u.Creci, &u.CreciEstado, &u.CEP, &u.Logradouro, &u.Numero, &u.Complemento, &u.Bairro, &u.Cidade, &u.Estado, &u.Telefone, &u.NumeroWhatsapp,
			&u.CriadoEm, &u.AtualizadoEm,
		)
		if err != nil {
			return nil, err
		}
		usuarios = append(usuarios, &u)
	}
	return usuarios, nil
}

func (r *usuarioPostgres) Update(ctx context.Context, u *domain.Usuario) error {
	query := `
		UPDATE usuarios SET 
			nome_completo = $1, telefone = $2, numero_whatsapp = $3, cpf = $4, rg = $5, nacionalidade = $6,
			estado_civil = $7, creci = $8, creci_estado = $9, cep = $10, logradouro = $11, numero = $12,
			complemento = $13, bairro = $14, cidade = $15, estado = $16, rg_estado = $17, rg_orgao_expedidor = $18, atualizado_em = now()
		WHERE id = $19
	`
	_, err := r.db.ExecContext(ctx, query,
		u.NomeCompleto, u.Telefone, u.NumeroWhatsapp, u.CPF, u.RG, u.Nacionalidade,
		u.EstadoCivil, u.Creci, u.CreciEstado, u.CEP, u.Logradouro, u.Numero,
		u.Complemento, u.Bairro, u.Cidade, u.Estado, u.RGEstado, u.RGOrgaoExpedidor,
		u.ID,
	)
	return err
}

func (r *usuarioPostgres) AtualizarSenha(ctx context.Context, usuarioID, senhaHash string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE usuarios SET senha_hash=$1, versao_sessao=versao_sessao+1, atualizado_em=NOW() WHERE id=$2`, senhaHash, usuarioID)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas != 1 {
		return errors.New("usuário não encontrado")
	}
	return nil
}

func (r *usuarioPostgres) VincularGoogle(ctx context.Context, usuarioID, googleID string, avatar *string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE usuarios SET google_id=$1, url_avatar=COALESCE($2, url_avatar), atualizado_em=NOW()
		WHERE id=$3 AND (google_id IS NULL OR google_id=$1)`, googleID, avatar, usuarioID)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas != 1 {
		return errors.New("não foi possível vincular a identidade Google")
	}
	return nil
}

func (r *usuarioPostgres) AtualizarStatusEquipe(ctx context.Context, usuarioID, contaID, status string, limiteCorretores *int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, contaID); err != nil {
		return err
	}
	if status == "ATIVO" && limiteCorretores != nil {
		var ativos int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM usuarios
			WHERE conta_id=$1 AND papel='CORRETOR_EQUIPE' AND UPPER(status)='ATIVO' AND id<>$2`, contaID, usuarioID).Scan(&ativos); err != nil {
			return err
		}
		if ativos >= *limiteCorretores {
			return errors.New("limite de corretores ativos do plano atingido")
		}
	}
	resultado, err := tx.ExecContext(ctx, `UPDATE usuarios
		SET status=$1, versao_sessao=versao_sessao+1, atualizado_em=NOW()
		WHERE id=$2 AND conta_id=$3 AND papel='CORRETOR_EQUIPE'`, status, usuarioID, contaID)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas != 1 {
		return errors.New("corretor da equipe não encontrado")
	}
	return tx.Commit()
}
