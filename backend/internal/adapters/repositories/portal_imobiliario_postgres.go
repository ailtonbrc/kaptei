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

type portalImobiliarioPostgres struct{ db *sql.DB }

func NewPortalImobiliarioPostgres(db *sql.DB) ports.PortalImobiliarioRepository {
	return &portalImobiliarioPostgres{db: db}
}

func (r *portalImobiliarioPostgres) ObterConfiguracao(ctx context.Context, contaID, portal string) (*domain.ConfiguracaoPortal, error) {
	configuracao := &domain.ConfiguracaoPortal{ContaID: contaID, Portal: portal, ExibicaoEndereco: "BAIRRO"}
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(i.id::text,''),COALESCE(i.ativa,false),i.token_feed_prefixo,
		COALESCE(NULLIF(i.nome_contato,''),c.nome_conta,''),COALESCE(NULLIF(i.email_contato,''),c.site_config->>'email',''),
		COALESCE(NULLIF(i.telefone_contato,''),c.site_config->>'telefone',''),COALESCE(i.exibicao_endereco,'BAIRRO'),
		COALESCE(i.atualizado_em,c.atualizado_em)
		FROM contas_saas c LEFT JOIN integracoes_portais i ON i.conta_id=c.id AND i.portal=$2
		WHERE c.id=$1`, contaID, portal).Scan(&configuracao.ID, &configuracao.Ativa, &configuracao.TokenFeedPrefixo,
		&configuracao.NomeContato, &configuracao.EmailContato, &configuracao.TelefoneContato,
		&configuracao.ExibicaoEndereco, &configuracao.AtualizadoEm)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("conta não encontrada")
	}
	if err != nil {
		return nil, fmt.Errorf("obter configuração do portal: %w", err)
	}
	return configuracao, nil
}

func (r *portalImobiliarioPostgres) SalvarConfiguracao(ctx context.Context, configuracao *domain.ConfiguracaoPortal, usuarioID string) error {
	resultado, err := r.db.ExecContext(ctx, `INSERT INTO integracoes_portais
		(conta_id,portal,ativa,nome_contato,email_contato,telefone_contato,exibicao_endereco,atualizado_por)
		SELECT id,$2,$3,$4,$5,$6,$7,$8 FROM contas_saas WHERE id=$1
		ON CONFLICT (conta_id,portal) DO UPDATE SET ativa=EXCLUDED.ativa,nome_contato=EXCLUDED.nome_contato,
		email_contato=EXCLUDED.email_contato,telefone_contato=EXCLUDED.telefone_contato,
		exibicao_endereco=EXCLUDED.exibicao_endereco,atualizado_por=EXCLUDED.atualizado_por,atualizado_em=now()`,
		configuracao.ContaID, configuracao.Portal, configuracao.Ativa, configuracao.NomeContato,
		configuracao.EmailContato, configuracao.TelefoneContato, configuracao.ExibicaoEndereco, usuarioID)
	if err != nil {
		return fmt.Errorf("salvar configuração do portal: %w", err)
	}
	return exigirLinhaAlterada(resultado, nil, "conta não encontrada")
}

func (r *portalImobiliarioPostgres) RotacionarToken(ctx context.Context, contaID, portal, hash, prefixo, usuarioID string) error {
	resultado, err := r.db.ExecContext(ctx, `INSERT INTO integracoes_portais
		(conta_id,portal,token_feed_hash,token_feed_prefixo,nome_contato,email_contato,telefone_contato,atualizado_por)
		SELECT c.id,$2,$3,$4,COALESCE(c.nome_conta,''),COALESCE(c.site_config->>'email',''),
		COALESCE(c.site_config->>'telefone',''),$5 FROM contas_saas c WHERE c.id=$1
		ON CONFLICT (conta_id,portal) DO UPDATE SET token_feed_hash=EXCLUDED.token_feed_hash,
		token_feed_prefixo=EXCLUDED.token_feed_prefixo,atualizado_por=EXCLUDED.atualizado_por,atualizado_em=now()`,
		contaID, portal, hash, prefixo, usuarioID)
	if err != nil {
		return fmt.Errorf("rotacionar token do feed: %w", err)
	}
	return exigirLinhaAlterada(resultado, nil, "conta não encontrada")
}

func (r *portalImobiliarioPostgres) ObterContaPorToken(ctx context.Context, portal, hash string) (string, error) {
	var contaID string
	err := r.db.QueryRowContext(ctx, `SELECT conta_id FROM integracoes_portais
		WHERE portal=$1 AND token_feed_hash=$2 AND ativa=true`, portal, hash).Scan(&contaID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("feed não encontrado")
	}
	if err != nil {
		return "", fmt.Errorf("resolver token do feed: %w", err)
	}
	return contaID, nil
}

func (r *portalImobiliarioPostgres) ListarPublicacoes(ctx context.Context, contaID, portal string) ([]domain.PublicacaoPortal, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT i.id,i.titulo,i.tipo,i.finalidade,i.status,
		COALESCE(p.ativa,false),COALESCE(p.tipo_publicacao,'STANDARD')
		FROM imoveis i LEFT JOIN publicacoes_portais p ON p.conta_id=i.conta_id AND p.imovel_id=i.id AND p.portal=$2
		WHERE i.conta_id=$1 ORDER BY i.atualizado_em DESC,i.id`, contaID, portal)
	if err != nil {
		return nil, fmt.Errorf("listar publicações do portal: %w", err)
	}
	defer rows.Close()
	publicacoes := make([]domain.PublicacaoPortal, 0)
	for rows.Next() {
		var publicacao domain.PublicacaoPortal
		if err := rows.Scan(&publicacao.ImovelID, &publicacao.Titulo, &publicacao.Tipo, &publicacao.Finalidade,
			&publicacao.Status, &publicacao.Ativa, &publicacao.TipoPublicacao); err != nil {
			return nil, err
		}
		publicacao.Erros = []string{}
		publicacoes = append(publicacoes, publicacao)
	}
	return publicacoes, rows.Err()
}

func (r *portalImobiliarioPostgres) SalvarPublicacao(ctx context.Context, contaID, portal, imovelID, usuarioID string, atualizacao domain.AtualizacaoPublicacaoPortal) error {
	resultado, err := r.db.ExecContext(ctx, `INSERT INTO publicacoes_portais
		(conta_id,portal,imovel_id,ativa,tipo_publicacao,atualizado_por)
		SELECT conta_id,$2,id,$4,$5,$6 FROM imoveis WHERE conta_id=$1 AND id=$3
		ON CONFLICT (conta_id,portal,imovel_id) DO UPDATE SET ativa=EXCLUDED.ativa,
		tipo_publicacao=EXCLUDED.tipo_publicacao,atualizado_por=EXCLUDED.atualizado_por,atualizado_em=now()`,
		contaID, portal, imovelID, atualizacao.Ativa, atualizacao.TipoPublicacao, usuarioID)
	if err != nil {
		return fmt.Errorf("salvar publicação do portal: %w", err)
	}
	return exigirLinhaAlterada(resultado, nil, "imóvel não encontrado nesta conta")
}

func (r *portalImobiliarioPostgres) ObterImovelDaConta(ctx context.Context, contaID, imovelID string) (*string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `SELECT i.id FROM imoveis i
		JOIN publicacoes_portais p ON p.conta_id=i.conta_id AND p.imovel_id=i.id
		WHERE i.id=$1 AND i.conta_id=$2 AND p.portal=$3 AND p.ativa=true`,
		imovelID, contaID, domain.PortalGrupoOLX).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("validar imóvel do lead: %w", err)
	}
	return &id, nil
}

func (r *portalImobiliarioPostgres) ObterDadosFeed(ctx context.Context, contaID, portal string) (*domain.DadosFeedPortal, error) {
	configuracao, err := r.ObterConfiguracao(ctx, contaID, portal)
	if err != nil {
		return nil, err
	}
	dados := &domain.DadosFeedPortal{Configuracao: *configuracao, Anuncios: make([]domain.AnuncioPortal, 0)}
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(nome_conta,''),COALESCE(slug_publico,''),site_publicado
		FROM contas_saas WHERE id=$1`, contaID).Scan(&dados.NomeConta, &dados.SiteSlug, &dados.SitePublicado); err != nil {
		return nil, fmt.Errorf("obter site para feed: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, selecionarAnunciosPortal, contaID, portal)
	if err != nil {
		return nil, fmt.Errorf("listar anúncios do feed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		anuncio, err := scanAnuncioPortal(rows)
		if err != nil {
			return nil, err
		}
		dados.Anuncios = append(dados.Anuncios, *anuncio)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("percorrer anúncios do feed: %w", err)
	}
	return dados, nil
}

const selecionarAnunciosPortal = `SELECT i.id,COALESCE(i.slug_publico,''),i.titulo,i.tipo,i.finalidade,i.status,
	i.valor_venda,i.valor_locacao,i.valor_condominio,i.valor_iptu,i.area_total,i.area_util,
	i.quartos,i.suites,i.banheiros,i.vagas,i.cep,i.logradouro,i.numero,i.complemento,i.bairro,i.cidade,i.estado,i.descricao,
	p.tipo_publicacao,COALESCE((SELECT JSON_AGG(JSON_BUILD_OBJECT(
	'id',f.id,'imovel_id',f.imovel_id,'url',f.url,'url_thumbnail',f.url_thumbnail,
	'tipo_conteudo',f.tipo_conteudo,'tamanho_bytes',f.tamanho_bytes,'largura',f.largura,'altura',f.altura,
	'hash_sha256',f.hash_sha256,'ordem',f.ordem,'is_capa',f.is_capa,'criado_em',f.criado_em
	) ORDER BY f.is_capa DESC,f.ordem,f.id) FROM imovel_fotos f WHERE f.imovel_id=i.id),'[]'::json)
	FROM publicacoes_portais p JOIN imoveis i ON i.id=p.imovel_id AND i.conta_id=p.conta_id
	WHERE p.conta_id=$1 AND p.portal=$2 AND p.ativa=true ORDER BY i.id`

func scanAnuncioPortal(origem scanner) (*domain.AnuncioPortal, error) {
	anuncio := &domain.AnuncioPortal{}
	var fotosJSON []byte
	err := origem.Scan(&anuncio.ID, &anuncio.Slug, &anuncio.Titulo, &anuncio.Tipo, &anuncio.Finalidade, &anuncio.Status,
		&anuncio.ValorVenda, &anuncio.ValorLocacao, &anuncio.ValorCondominio, &anuncio.ValorIPTU,
		&anuncio.AreaTotal, &anuncio.AreaUtil, &anuncio.Quartos, &anuncio.Suites, &anuncio.Banheiros, &anuncio.Vagas,
		&anuncio.CEP, &anuncio.Logradouro, &anuncio.Numero, &anuncio.Complemento, &anuncio.Bairro,
		&anuncio.Cidade, &anuncio.Estado, &anuncio.Descricao, &anuncio.TipoPublicacao, &fotosJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(fotosJSON, &anuncio.Fotos); err != nil {
		return nil, fmt.Errorf("decodificar fotos do anúncio: %w", err)
	}
	return anuncio, nil
}
