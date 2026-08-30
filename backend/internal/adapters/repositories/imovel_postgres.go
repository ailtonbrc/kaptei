package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/msdev/kaptei/internal/core/domain"
)

type imovelPostgres struct {
	db *sql.DB
}

func NewImovelPostgres(db *sql.DB) *imovelPostgres {
	return &imovelPostgres{db: db}
}

func (r *imovelPostgres) Create(ctx context.Context, imovel *domain.Imovel) error {
	query := `
		INSERT INTO imoveis (
			conta_id, usuario_id, titulo, tipo, finalidade, status,
			valor_venda, valor_locacao, valor_condominio, valor_iptu,
			area_total, area_util, quartos, suites, banheiros, vagas,
			cep, logradouro, numero, complemento, bairro, cidade, estado, descricao,
			slug_publico, publicado, destaque, titulo_seo, descricao_seo
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
		) RETURNING id, criado_em, atualizado_em
	`

	err := r.db.QueryRowContext(
		ctx, query,
		imovel.ContaID, imovel.UsuarioID, imovel.Titulo, imovel.Tipo, imovel.Finalidade, imovel.Status,
		imovel.ValorVenda, imovel.ValorLocacao, imovel.ValorCondominio, imovel.ValorIPTU,
		imovel.AreaTotal, imovel.AreaUtil, imovel.Quartos, imovel.Suites, imovel.Banheiros, imovel.Vagas,
		imovel.CEP, imovel.Logradouro, imovel.Numero, imovel.Complemento, imovel.Bairro, imovel.Cidade, imovel.Estado, imovel.Descricao,
		imovel.SlugPublico, imovel.Publicado, imovel.Destaque, imovel.TituloSEO, imovel.DescricaoSEO,
	).Scan(&imovel.ID, &imovel.CriadoEm, &imovel.AtualizadoEm)

	if err != nil {
		return fmt.Errorf("erro ao criar imóvel: %v", err)
	}

	return nil
}

func (r *imovelPostgres) GetByID(ctx context.Context, id, contaID string) (*domain.Imovel, error) {
	query := `
		SELECT 
			id, conta_id, usuario_id, titulo, tipo, finalidade, status,
			valor_venda, valor_locacao, valor_condominio, valor_iptu,
			area_total, area_util, quartos, suites, banheiros, vagas,
			cep, logradouro, numero, complemento, bairro, cidade, estado, descricao,
			slug_publico, publicado, destaque, titulo_seo, descricao_seo,
			criado_em, atualizado_em
		FROM imoveis
		WHERE id = $1 AND conta_id = $2
	`

	imovel := &domain.Imovel{}
	err := r.db.QueryRowContext(ctx, query, id, contaID).Scan(
		&imovel.ID, &imovel.ContaID, &imovel.UsuarioID, &imovel.Titulo, &imovel.Tipo, &imovel.Finalidade, &imovel.Status,
		&imovel.ValorVenda, &imovel.ValorLocacao, &imovel.ValorCondominio, &imovel.ValorIPTU,
		&imovel.AreaTotal, &imovel.AreaUtil, &imovel.Quartos, &imovel.Suites, &imovel.Banheiros, &imovel.Vagas,
		&imovel.CEP, &imovel.Logradouro, &imovel.Numero, &imovel.Complemento, &imovel.Bairro, &imovel.Cidade, &imovel.Estado, &imovel.Descricao,
		&imovel.SlugPublico, &imovel.Publicado, &imovel.Destaque, &imovel.TituloSEO, &imovel.DescricaoSEO,
		&imovel.CriadoEm, &imovel.AtualizadoEm,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Não encontrado
		}
		return nil, fmt.Errorf("erro ao buscar imóvel: %v", err)
	}

	// Buscar as fotos
	fotosQuery := `SELECT id,imovel_id,url,url_thumbnail,chave_objeto,chave_thumbnail,provedor_storage,
		tipo_conteudo,tamanho_bytes,largura,altura,hash_sha256,ordem,is_capa,criado_em
		FROM imovel_fotos WHERE imovel_id = $1 ORDER BY ordem ASC`
	rows, err := r.db.QueryContext(ctx, fotosQuery, imovel.ID)
	if err != nil {
		return nil, fmt.Errorf("carregar fotos do imóvel: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var foto domain.ImovelFoto
		if err := scanFoto(rows, &foto); err != nil {
			return nil, fmt.Errorf("ler foto do imóvel: %w", err)
		}
		imovel.Fotos = append(imovel.Fotos, foto)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("percorrer fotos do imóvel: %w", err)
	}

	return imovel, nil
}

func (r *imovelPostgres) ListByContaID(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Imovel], error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT 
			id, conta_id, usuario_id, titulo, tipo, finalidade, status,
			valor_venda, valor_locacao, valor_condominio, valor_iptu,
			area_total, area_util, quartos, suites, banheiros, vagas,
			cep, logradouro, numero, complemento, bairro, cidade, estado, descricao,
			slug_publico, publicado, destaque, titulo_seo, descricao_seo,
			criado_em, atualizado_em
		FROM imoveis
		WHERE conta_id = $1
	`)

	args := []interface{}{contaID}
	argIdx := 2

	if filtro.Tipo != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND tipo = $%d", argIdx))
		args = append(args, filtro.Tipo)
		argIdx++
	}
	if filtro.Finalidade != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND finalidade = $%d", argIdx))
		args = append(args, filtro.Finalidade)
		argIdx++
	}
	if filtro.Status != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argIdx))
		args = append(args, filtro.Status)
		argIdx++
	}
	if filtro.Busca != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND (titulo ILIKE '%%' || $%d || '%%' OR COALESCE(bairro, '') ILIKE '%%' || $%d || '%%' OR COALESCE(cidade, '') ILIKE '%%' || $%d || '%%')", argIdx, argIdx, argIdx))
		args = append(args, filtro.Busca)
		argIdx++
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ("+queryBuilder.String()+") AS imoveis_filtrados", args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("contar imóveis: %w", err)
	}
	deslocamento := (filtro.Pagina - 1) * filtro.Limite
	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY criado_em DESC, id DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1))
	args = append(args, filtro.Limite, deslocamento)

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar imóveis: %v", err)
	}
	defer rows.Close()

	var imoveis []*domain.Imovel
	for rows.Next() {
		imovel := &domain.Imovel{}
		err := rows.Scan(
			&imovel.ID, &imovel.ContaID, &imovel.UsuarioID, &imovel.Titulo, &imovel.Tipo, &imovel.Finalidade, &imovel.Status,
			&imovel.ValorVenda, &imovel.ValorLocacao, &imovel.ValorCondominio, &imovel.ValorIPTU,
			&imovel.AreaTotal, &imovel.AreaUtil, &imovel.Quartos, &imovel.Suites, &imovel.Banheiros, &imovel.Vagas,
			&imovel.CEP, &imovel.Logradouro, &imovel.Numero, &imovel.Complemento, &imovel.Bairro, &imovel.Cidade, &imovel.Estado, &imovel.Descricao,
			&imovel.SlugPublico, &imovel.Publicado, &imovel.Destaque, &imovel.TituloSEO, &imovel.DescricaoSEO,
			&imovel.CriadoEm, &imovel.AtualizadoEm,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao fazer scan de imóvel: %v", err)
		}
		imoveis = append(imoveis, imovel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("percorrer imóveis: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("encerrar consulta de imóveis: %w", err)
	}

	if len(imoveis) > 0 {
		ids := make([]string, 0, len(imoveis))
		porID := make(map[string]*domain.Imovel, len(imoveis))
		for _, imovel := range imoveis {
			ids = append(ids, imovel.ID)
			porID[imovel.ID] = imovel
		}
		fotos, err := r.db.QueryContext(ctx, `SELECT DISTINCT ON (imovel_id)
			id,imovel_id,url,url_thumbnail,chave_objeto,chave_thumbnail,provedor_storage,
			tipo_conteudo,tamanho_bytes,largura,altura,hash_sha256,ordem,is_capa,criado_em FROM imovel_fotos
			WHERE imovel_id = ANY($1) ORDER BY imovel_id, is_capa DESC, ordem, criado_em, id`, pq.Array(ids))
		if err != nil {
			return nil, fmt.Errorf("carregar capas dos imóveis: %w", err)
		}
		defer fotos.Close()
		for fotos.Next() {
			var foto domain.ImovelFoto
			if err := scanFoto(fotos, &foto); err != nil {
				return nil, fmt.Errorf("ler capa do imóvel: %w", err)
			}
			porID[foto.ImovelID].Fotos = []domain.ImovelFoto{foto}
		}
		if err := fotos.Err(); err != nil {
			return nil, fmt.Errorf("percorrer capas dos imóveis: %w", err)
		}
	}

	if imoveis == nil {
		imoveis = []*domain.Imovel{}
	}
	return &domain.ListaPaginada[*domain.Imovel]{Dados: imoveis, Total: total, Pagina: filtro.Pagina, Limite: filtro.Limite}, nil
}

func (r *imovelPostgres) Update(ctx context.Context, imovel *domain.Imovel) error {
	query := `
		UPDATE imoveis SET
			titulo = $1, tipo = $2, finalidade = $3, status = $4,
			valor_venda = $5, valor_locacao = $6, valor_condominio = $7, valor_iptu = $8,
			area_total = $9, area_util = $10, quartos = $11, suites = $12, banheiros = $13, vagas = $14,
			cep = $15, logradouro = $16, numero = $17, complemento = $18, bairro = $19, cidade = $20, estado = $21, descricao = $22,
			slug_publico = $23, publicado = $24, destaque = $25, titulo_seo = $26, descricao_seo = $27,
			atualizado_em = CURRENT_TIMESTAMP
		WHERE id = $28 AND conta_id = $29
	`

	res, err := r.db.ExecContext(
		ctx, query,
		imovel.Titulo, imovel.Tipo, imovel.Finalidade, imovel.Status,
		imovel.ValorVenda, imovel.ValorLocacao, imovel.ValorCondominio, imovel.ValorIPTU,
		imovel.AreaTotal, imovel.AreaUtil, imovel.Quartos, imovel.Suites, imovel.Banheiros, imovel.Vagas,
		imovel.CEP, imovel.Logradouro, imovel.Numero, imovel.Complemento, imovel.Bairro, imovel.Cidade, imovel.Estado, imovel.Descricao,
		imovel.SlugPublico, imovel.Publicado, imovel.Destaque, imovel.TituloSEO, imovel.DescricaoSEO,
		imovel.ID, imovel.ContaID,
	)

	if err != nil {
		return fmt.Errorf("erro ao atualizar imóvel: %v", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("imóvel não encontrado ou sem permissão para atualizar")
	}

	return nil
}

func (r *imovelPostgres) Delete(ctx context.Context, id, contaID string, eventos []*domain.EventoOutbox) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar exclusão do imóvel: %w", err)
	}
	defer tx.Rollback()
	query := `DELETE FROM imoveis WHERE id = $1 AND conta_id = $2`
	res, err := tx.ExecContext(ctx, query, id, contaID)
	if err != nil {
		return fmt.Errorf("erro ao excluir imóvel: %v", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("imóvel não encontrado ou sem permissão para excluir")
	}
	for _, evento := range eventos {
		if err := inserirEventoOutbox(ctx, tx, evento); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *imovelPostgres) AddFoto(ctx context.Context, foto *domain.ImovelFoto) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar inclusão da foto: %w", err)
	}
	defer tx.Rollback()
	var imovelExiste int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM imoveis WHERE id=$1 FOR UPDATE`, foto.ImovelID).Scan(&imovelExiste); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("imóvel não encontrado")
		}
		return fmt.Errorf("bloquear imóvel para incluir foto: %w", err)
	}
	var quantidade int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(MAX(ordem), -1) + 1 FROM imovel_fotos WHERE imovel_id=$1`, foto.ImovelID).Scan(&quantidade, &foto.Ordem); err != nil {
		return fmt.Errorf("calcular ordem da foto: %w", err)
	}
	if quantidade >= 50 {
		return errors.New("o imóvel atingiu o limite de 50 fotos")
	}
	if quantidade == 0 {
		foto.IsCapa = true
	}
	if foto.IsCapa {
		if _, err := tx.ExecContext(ctx, `UPDATE imovel_fotos SET is_capa=FALSE WHERE imovel_id=$1 AND is_capa=TRUE`, foto.ImovelID); err != nil {
			return fmt.Errorf("substituir capa do imóvel: %w", err)
		}
	}
	query := `
		INSERT INTO imovel_fotos (
			imovel_id,url,url_thumbnail,chave_objeto,chave_thumbnail,provedor_storage,
			tipo_conteudo,tamanho_bytes,largura,altura,hash_sha256,ordem,is_capa
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, criado_em
	`
	err = tx.QueryRowContext(ctx, query,
		foto.ImovelID, foto.URL, foto.URLThumbnail, foto.ChaveObjeto, foto.ChaveThumbnail,
		foto.ProvedorStorage, foto.TipoConteudo, foto.TamanhoBytes, foto.Largura, foto.Altura,
		foto.HashSHA256, foto.Ordem, foto.IsCapa,
	).Scan(&foto.ID, &foto.CriadoEm)
	if err != nil {
		return fmt.Errorf("adicionar foto: %w", err)
	}
	return tx.Commit()
}

func (r *imovelPostgres) GetFoto(ctx context.Context, fotoID, imovelID string) (*domain.ImovelFoto, error) {
	foto := &domain.ImovelFoto{}
	err := scanFoto(r.db.QueryRowContext(ctx, `SELECT id,imovel_id,url,url_thumbnail,chave_objeto,chave_thumbnail,provedor_storage,
		tipo_conteudo,tamanho_bytes,largura,altura,hash_sha256,ordem,is_capa,criado_em
		FROM imovel_fotos WHERE id=$1 AND imovel_id=$2`, fotoID, imovelID), foto)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("carregar foto: %w", err)
	}
	return foto, nil
}

func (r *imovelPostgres) DeleteFoto(ctx context.Context, fotoID, imovelID string, eventos []*domain.EventoOutbox) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar exclusão da foto: %w", err)
	}
	defer tx.Rollback()
	var eraCapa bool
	err = tx.QueryRowContext(ctx, `DELETE FROM imovel_fotos WHERE id=$1 AND imovel_id=$2 RETURNING is_capa`, fotoID, imovelID).Scan(&eraCapa)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("foto não encontrada")
	}
	if err != nil {
		return fmt.Errorf("excluir foto: %w", err)
	}
	if eraCapa {
		if _, err := tx.ExecContext(ctx, `UPDATE imovel_fotos SET is_capa=TRUE WHERE id=(
			SELECT id FROM imovel_fotos WHERE imovel_id=$1 ORDER BY ordem, criado_em, id LIMIT 1
		)`, imovelID); err != nil {
			return fmt.Errorf("promover nova capa: %w", err)
		}
	}
	for _, evento := range eventos {
		if err := inserirEventoOutbox(ctx, tx, evento); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func scanFoto(origem interface{ Scan(...any) error }, foto *domain.ImovelFoto) error {
	return origem.Scan(
		&foto.ID, &foto.ImovelID, &foto.URL, &foto.URLThumbnail,
		&foto.ChaveObjeto, &foto.ChaveThumbnail, &foto.ProvedorStorage,
		&foto.TipoConteudo, &foto.TamanhoBytes, &foto.Largura, &foto.Altura, &foto.HashSHA256,
		&foto.Ordem, &foto.IsCapa, &foto.CriadoEm,
	)
}
