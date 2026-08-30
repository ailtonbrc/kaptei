package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/msdev/kaptei/internal/core/domain"
)

type sitePublicoPostgres struct{ db *sql.DB }

func NewSitePublicoPostgres(db *sql.DB) *sitePublicoPostgres {
	return &sitePublicoPostgres{db: db}
}

func (r *sitePublicoPostgres) GetBySlug(ctx context.Context, slug string) (*domain.SitePublico, error) {
	return r.buscarSite(ctx, `LOWER(slug_publico) = LOWER($1) AND site_publicado = TRUE`, slug)
}

func (r *sitePublicoPostgres) GetByContaID(ctx context.Context, contaID string) (*domain.SitePublico, error) {
	return r.buscarSite(ctx, `id = $1`, contaID)
}

func (r *sitePublicoPostgres) buscarSite(ctx context.Context, condicao, valor string) (*domain.SitePublico, error) {
	query := `SELECT id, COALESCE(slug_publico, ''), COALESCE(nome_conta, ''), site_publicado, site_config
		FROM contas_saas WHERE ` + condicao
	site := &domain.SitePublico{}
	var configuracaoJSON []byte
	if err := r.db.QueryRowContext(ctx, query, valor).Scan(
		&site.ContaID, &site.Slug, &site.Nome, &site.Publicado, &configuracaoJSON,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("buscar site público: %w", err)
	}
	if err := json.Unmarshal(configuracaoJSON, &site.Configuracao); err != nil {
		return nil, fmt.Errorf("decodificar configuração do site: %w", err)
	}
	return site, nil
}

func (r *sitePublicoPostgres) Salvar(ctx context.Context, site *domain.SitePublico) error {
	configuracaoJSON, err := json.Marshal(site.Configuracao)
	if err != nil {
		return fmt.Errorf("codificar configuração do site: %w", err)
	}
	resultado, err := r.db.ExecContext(ctx, `
		UPDATE contas_saas
		SET slug_publico = NULLIF($1, ''), site_publicado = $2, site_config = $3, atualizado_em = NOW()
		WHERE id = $4`, site.Slug, site.Publicado, configuracaoJSON, site.ContaID)
	if err != nil {
		return fmt.Errorf("salvar site público: %w", err)
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return errors.New("conta não encontrada para configurar o site")
	}
	return nil
}

func (r *sitePublicoPostgres) ListarImoveis(ctx context.Context, contaID string, filtros domain.FiltrosCatalogoPublico) ([]*domain.ImovelPublico, int, error) {
	where, args := montarFiltrosCatalogo(contaID, filtros)
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM imoveis i `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar imóveis públicos: %w", err)
	}

	args = append(args, filtros.Limite, (filtros.Pagina-1)*filtros.Limite)
	query := selecionarImovelPublico + where + fmt.Sprintf(
		" ORDER BY i.destaque DESC, i.atualizado_em DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args),
	)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar imóveis públicos: %w", err)
	}
	defer rows.Close()

	imoveis := make([]*domain.ImovelPublico, 0)
	for rows.Next() {
		imovel, err := scanImovelPublico(rows)
		if err != nil {
			return nil, 0, err
		}
		imoveis = append(imoveis, imovel)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("percorrer imóveis públicos: %w", err)
	}
	return imoveis, total, nil
}

func (r *sitePublicoPostgres) GetImovelBySlug(ctx context.Context, contaID, slug string) (*domain.ImovelPublico, error) {
	row := r.db.QueryRowContext(ctx, selecionarImovelPublico+`
		WHERE i.conta_id = $1 AND LOWER(i.slug_publico) = LOWER($2)
		AND i.publicado = TRUE AND i.status = 'Ativo'`, contaID, slug)
	imovel, err := scanImovelPublico(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return imovel, err
}

func (r *sitePublicoPostgres) ListarRotasSitemap(ctx context.Context) ([]domain.RotaSitemap, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.slug_publico, i.slug_publico, GREATEST(c.atualizado_em, COALESCE(i.atualizado_em, c.atualizado_em))
		FROM contas_saas c
		LEFT JOIN imoveis i ON i.conta_id = c.id AND i.publicado = TRUE AND i.status = 'Ativo' AND i.slug_publico IS NOT NULL
		WHERE c.site_publicado = TRUE AND c.slug_publico IS NOT NULL
		ORDER BY c.slug_publico, i.slug_publico`)
	if err != nil {
		return nil, fmt.Errorf("listar rotas do sitemap: %w", err)
	}
	defer rows.Close()
	rotas := make([]domain.RotaSitemap, 0)
	for rows.Next() {
		var rota domain.RotaSitemap
		if err := rows.Scan(&rota.SlugSite, &rota.SlugImovel, &rota.AtualizadoEm); err != nil {
			return nil, fmt.Errorf("ler rota do sitemap: %w", err)
		}
		rotas = append(rotas, rota)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("percorrer rotas do sitemap: %w", err)
	}
	return rotas, nil
}

const selecionarImovelPublico = `
	SELECT i.id, COALESCE(i.slug_publico, ''), i.titulo, i.tipo, i.finalidade,
		i.valor_venda, i.valor_locacao, i.valor_condominio, i.valor_iptu,
		i.area_total, i.area_util, i.quartos, i.suites, i.banheiros, i.vagas,
		i.bairro, i.cidade, i.estado, i.descricao, i.titulo_seo, i.descricao_seo, i.destaque,
		COALESCE((
			SELECT JSON_AGG(JSON_BUILD_OBJECT(
				'id', f.id, 'imovel_id', f.imovel_id, 'url', f.url,
				'url_thumbnail', f.url_thumbnail, 'tipo_conteudo', f.tipo_conteudo,
				'tamanho_bytes', f.tamanho_bytes, 'largura', f.largura,
				'altura', f.altura, 'hash_sha256', f.hash_sha256,
				'ordem', f.ordem, 'is_capa', f.is_capa, 'criado_em', f.criado_em
			) ORDER BY f.is_capa DESC, f.ordem ASC)
			FROM imovel_fotos f WHERE f.imovel_id = i.id
		), '[]'::json)
	FROM imoveis i `

type scanner interface{ Scan(dest ...any) error }

func scanImovelPublico(origem scanner) (*domain.ImovelPublico, error) {
	imovel := &domain.ImovelPublico{}
	var fotosJSON []byte
	err := origem.Scan(
		&imovel.ID, &imovel.Slug, &imovel.Titulo, &imovel.Tipo, &imovel.Finalidade,
		&imovel.ValorVenda, &imovel.ValorLocacao, &imovel.ValorCondominio, &imovel.ValorIPTU,
		&imovel.AreaTotal, &imovel.AreaUtil, &imovel.Quartos, &imovel.Suites, &imovel.Banheiros, &imovel.Vagas,
		&imovel.Bairro, &imovel.Cidade, &imovel.Estado, &imovel.Descricao, &imovel.TituloSEO, &imovel.DescricaoSEO,
		&imovel.Destaque, &fotosJSON,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(fotosJSON, &imovel.Fotos); err != nil {
		return nil, fmt.Errorf("decodificar fotos públicas: %w", err)
	}
	return imovel, nil
}

func montarFiltrosCatalogo(contaID string, filtros domain.FiltrosCatalogoPublico) (string, []any) {
	var clausulas = []string{"i.conta_id = $1", "i.publicado = TRUE", "i.status = 'Ativo'"}
	args := []any{contaID}
	adicionar := func(expressao string, valor any) {
		args = append(args, valor)
		clausulas = append(clausulas, fmt.Sprintf(expressao, len(args)))
	}
	if filtros.Tipo != "" {
		adicionar("LOWER(i.tipo) = LOWER($%d)", filtros.Tipo)
	}
	if filtros.Finalidade != "" {
		adicionar("LOWER(i.finalidade) = LOWER($%d)", filtros.Finalidade)
	}
	if filtros.Cidade != "" {
		adicionar("LOWER(i.cidade) = LOWER($%d)", filtros.Cidade)
	}
	if filtros.Bairro != "" {
		adicionar("LOWER(i.bairro) = LOWER($%d)", filtros.Bairro)
	}
	if filtros.QuartosMin != nil {
		adicionar("i.quartos >= $%d", *filtros.QuartosMin)
	}
	if filtros.ValorMin != nil {
		adicionar("COALESCE(i.valor_venda, i.valor_locacao) >= $%d", *filtros.ValorMin)
	}
	if filtros.ValorMax != nil {
		adicionar("COALESCE(i.valor_venda, i.valor_locacao) <= $%d", *filtros.ValorMax)
	}
	return "WHERE " + strings.Join(clausulas, " AND "), args
}
