package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type dominioSitePostgres struct{ db *sql.DB }

func NewDominioSitePostgres(db *sql.DB) ports.DominioSiteRepository {
	return &dominioSitePostgres{db: db}
}

const camposDominioSite = `id,conta_id,hostname,token_verificacao,status,verificado_em,ultima_verificacao_em,ultimo_erro,criado_em,atualizado_em`

func escanearDominioSite(scanner interface{ Scan(...any) error }) (*domain.DominioSite, error) {
	var dominio domain.DominioSite
	err := scanner.Scan(&dominio.ID, &dominio.ContaID, &dominio.Hostname, &dominio.TokenVerificacao, &dominio.Status,
		&dominio.VerificadoEm, &dominio.UltimaVerificacaoEm, &dominio.UltimoErro, &dominio.CriadoEm, &dominio.AtualizadoEm)
	return &dominio, err
}

func (r *dominioSitePostgres) ObterPorConta(ctx context.Context, contaID string) (*domain.DominioSite, error) {
	dominio, err := escanearDominioSite(r.db.QueryRowContext(ctx, `SELECT `+camposDominioSite+` FROM dominios_site WHERE conta_id=$1`, contaID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return dominio, err
}

func (r *dominioSitePostgres) SalvarPendente(ctx context.Context, dominio *domain.DominioSite) error {
	err := r.db.QueryRowContext(ctx, `INSERT INTO dominios_site (conta_id,hostname,token_verificacao,status)
		VALUES ($1,$2,$3,'PENDENTE') ON CONFLICT (conta_id) DO UPDATE SET hostname=EXCLUDED.hostname,
		token_verificacao=EXCLUDED.token_verificacao,status='PENDENTE',verificado_em=NULL,
		ultima_verificacao_em=NULL,ultimo_erro=NULL,atualizado_em=now()
		RETURNING `+camposDominioSite, dominio.ContaID, dominio.Hostname, dominio.TokenVerificacao).Scan(
		&dominio.ID, &dominio.ContaID, &dominio.Hostname, &dominio.TokenVerificacao, &dominio.Status,
		&dominio.VerificadoEm, &dominio.UltimaVerificacaoEm, &dominio.UltimoErro, &dominio.CriadoEm, &dominio.AtualizadoEm,
	)
	if err != nil {
		return fmt.Errorf("salvar domínio do site: %w", err)
	}
	return nil
}

func (r *dominioSitePostgres) Ativar(ctx context.Context, id, contaID, tokenVerificacao string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE dominios_site SET status='ATIVO',verificado_em=now(),ultima_verificacao_em=now(),ultimo_erro=NULL,atualizado_em=now() WHERE id=$1 AND conta_id=$2 AND token_verificacao=$3`, id, contaID, tokenVerificacao)
	return exigirDominioAlterado(resultado, err)
}

func (r *dominioSitePostgres) RegistrarFalha(ctx context.Context, id, contaID, tokenVerificacao, mensagem string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE dominios_site SET status='FALHOU',ultima_verificacao_em=now(),ultimo_erro=$1,atualizado_em=now() WHERE id=$2 AND conta_id=$3 AND token_verificacao=$4`, mensagem, id, contaID, tokenVerificacao)
	return exigirDominioAlterado(resultado, err)
}

func (r *dominioSitePostgres) ObterSitePorHostname(ctx context.Context, hostname string) (*domain.SitePublico, error) {
	return r.buscarSiteDominio(ctx, `LOWER(d.hostname)=LOWER($1) AND d.status='ATIVO' AND c.site_publicado=TRUE`, hostname)
}

func (r *dominioSitePostgres) buscarSiteDominio(ctx context.Context, condicao, valor string) (*domain.SitePublico, error) {
	repositorioSite := &sitePublicoPostgres{db: r.db}
	var contaID string
	err := r.db.QueryRowContext(ctx, `SELECT c.id FROM dominios_site d JOIN contas_saas c ON c.id=d.conta_id WHERE `+condicao, valor).Scan(&contaID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("resolver domínio do site: %w", err)
	}
	return repositorioSite.GetByContaID(ctx, contaID)
}

func exigirDominioAlterado(resultado sql.Result, err error) error {
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas != 1 {
		return errors.New("domínio não encontrado")
	}
	return nil
}
