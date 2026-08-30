package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type integracaoMetaPostgres struct {
	db *sql.DB
}

func NewIntegracaoMetaPostgres(db *sql.DB) ports.IntegracaoMetaRepository {
	return &integracaoMetaPostgres{db: db}
}

func (r *integracaoMetaPostgres) ObterPorConta(ctx context.Context, contaID string) (*domain.ConfiguracaoMetaLeads, error) {
	return r.obter(ctx, `SELECT id,conta_id,pagina_id,token_pagina_protegido,ativa,criado_em,atualizado_em
		FROM integracoes_meta_leads WHERE conta_id=$1`, contaID)
}

func (r *integracaoMetaPostgres) ObterPorPagina(ctx context.Context, paginaID string) (*domain.ConfiguracaoMetaLeads, error) {
	return r.obter(ctx, `SELECT id,conta_id,pagina_id,token_pagina_protegido,ativa,criado_em,atualizado_em
		FROM integracoes_meta_leads WHERE pagina_id=$1 AND ativa=true`, paginaID)
}

func (r *integracaoMetaPostgres) obter(ctx context.Context, consulta, parametro string) (*domain.ConfiguracaoMetaLeads, error) {
	configuracao := &domain.ConfiguracaoMetaLeads{}
	err := r.db.QueryRowContext(ctx, consulta, parametro).Scan(
		&configuracao.ID, &configuracao.ContaID, &configuracao.PaginaID,
		&configuracao.TokenPaginaProtegido, &configuracao.Ativa,
		&configuracao.CriadoEm, &configuracao.AtualizadoEm,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("obter configuraÃ§Ã£o Meta Leads: %w", err)
	}
	configuracao.TokenPaginaConfigurado = configuracao.TokenPaginaProtegido != ""
	return configuracao, nil
}

func (r *integracaoMetaPostgres) Salvar(ctx context.Context, configuracao *domain.ConfiguracaoMetaLeads) error {
	err := r.db.QueryRowContext(ctx, `INSERT INTO integracoes_meta_leads
		(conta_id,pagina_id,token_pagina_protegido,ativa,criado_em,atualizado_em)
		VALUES ($1,$2,$3,$4,NOW(),NOW())
		ON CONFLICT (conta_id) DO UPDATE SET
			pagina_id=EXCLUDED.pagina_id,
			token_pagina_protegido=EXCLUDED.token_pagina_protegido,
			ativa=EXCLUDED.ativa,
			atualizado_em=NOW()
		RETURNING id,criado_em,atualizado_em`,
		configuracao.ContaID, configuracao.PaginaID, configuracao.TokenPaginaProtegido, configuracao.Ativa,
	).Scan(&configuracao.ID, &configuracao.CriadoEm, &configuracao.AtualizadoEm)
	if err != nil {
		var erroPostgres *pq.Error
		if errors.As(err, &erroPostgres) && erroPostgres.Code == "23505" {
			return errors.New("esta pÃ¡gina Meta jÃ¡ estÃ¡ vinculada a outra conta")
		}
		return fmt.Errorf("salvar configuraÃ§Ã£o Meta Leads: %w", err)
	}
	configuracao.TokenPaginaConfigurado = true
	return nil
}

func (r *integracaoMetaPostgres) Enfileirar(ctx context.Context, eventos []*domain.EventoIntegracao) error {
	if len(eventos) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar enfileiramento Meta: %w", err)
	}
	defer tx.Rollback()
	for _, evento := range eventos {
		if evento == nil {
			continue
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO eventos_integracao
			(conta_id,provedor,tipo,identificador_externo,pagina_id,formulario_id,anuncio_id,ocorrido_em,
			 status,tentativas,maximo_tentativas,disponivel_em,criado_em)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'PENDENTE',0,$9,$10,$11)
			ON CONFLICT (provedor,identificador_externo) DO NOTHING`,
			evento.ContaID, evento.Provedor, evento.Tipo, evento.IdentificadorExterno,
			evento.PaginaID, evento.FormularioID, evento.AnuncioID, evento.OcorridoEm,
			evento.MaximoTentativas, evento.DisponivelEm, evento.CriadoEm,
		)
		if err != nil {
			return fmt.Errorf("enfileirar evento Meta: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar enfileiramento Meta: %w", err)
	}
	return nil
}

func (r *integracaoMetaPostgres) Reservar(ctx context.Context, trabalhadorID string, limite int, duracaoBloqueio time.Duration) ([]*domain.EventoIntegracao, error) {
	rows, err := r.db.QueryContext(ctx, `WITH candidatos AS (
		SELECT id FROM eventos_integracao
		WHERE provedor='META' AND tipo='LEAD_GERADO' AND tentativas < maximo_tentativas
		  AND ((status='PENDENTE' AND disponivel_em<=NOW()) OR (status='PROCESSANDO' AND bloqueado_ate<NOW()))
		ORDER BY disponivel_em,criado_em FOR UPDATE SKIP LOCKED LIMIT $1
	)
	UPDATE eventos_integracao evento
	SET status='PROCESSANDO',tentativas=tentativas+1,bloqueado_por=$2,
		bloqueado_ate=NOW()+($3 * INTERVAL '1 millisecond')
	FROM candidatos WHERE evento.id=candidatos.id
	RETURNING evento.id,evento.conta_id,evento.provedor,evento.tipo,evento.identificador_externo,
		evento.pagina_id,evento.formulario_id,evento.anuncio_id,evento.ocorrido_em,evento.status,
		evento.tentativas,evento.maximo_tentativas,evento.disponivel_em,evento.bloqueado_ate,
		evento.bloqueado_por,evento.criado_em`, limite, trabalhadorID, duracaoBloqueio.Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("reservar eventos Meta: %w", err)
	}
	defer rows.Close()
	eventos := make([]*domain.EventoIntegracao, 0, limite)
	for rows.Next() {
		evento := &domain.EventoIntegracao{}
		if err := rows.Scan(
			&evento.ID, &evento.ContaID, &evento.Provedor, &evento.Tipo, &evento.IdentificadorExterno,
			&evento.PaginaID, &evento.FormularioID, &evento.AnuncioID, &evento.OcorridoEm, &evento.Status,
			&evento.Tentativas, &evento.MaximoTentativas, &evento.DisponivelEm, &evento.BloqueadoAte,
			&evento.BloqueadoPor, &evento.CriadoEm,
		); err != nil {
			return nil, fmt.Errorf("ler evento Meta reservado: %w", err)
		}
		eventos = append(eventos, evento)
	}
	return eventos, rows.Err()
}

func (r *integracaoMetaPostgres) Concluir(ctx context.Context, eventoID, trabalhadorID string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE eventos_integracao
		SET status='CONCLUIDO',processado_em=NOW(),bloqueado_ate=NULL,bloqueado_por=NULL,ultimo_erro=NULL
		WHERE id=$1 AND status='PROCESSANDO' AND bloqueado_por=$2`, eventoID, trabalhadorID)
	return validarAtualizacaoEventoIntegracao(resultado, err, "concluir")
}

func (r *integracaoMetaPostgres) Falhar(ctx context.Context, eventoID, trabalhadorID, mensagem string, proximaTentativa time.Time, definitivo bool) error {
	status := "PENDENTE"
	if definitivo {
		status = "FALHOU"
	}
	mensagem = strings.TrimSpace(mensagem)
	if len([]rune(mensagem)) > 1000 {
		mensagem = string([]rune(mensagem)[:1000])
	}
	resultado, err := r.db.ExecContext(ctx, `UPDATE eventos_integracao
		SET status=$3,disponivel_em=$4,bloqueado_ate=NULL,bloqueado_por=NULL,ultimo_erro=$5
		WHERE id=$1 AND status='PROCESSANDO' AND bloqueado_por=$2`,
		eventoID, trabalhadorID, status, proximaTentativa, mensagem)
	return validarAtualizacaoEventoIntegracao(resultado, err, "registrar falha de")
}

func validarAtualizacaoEventoIntegracao(resultado sql.Result, err error, acao string) error {
	if err != nil {
		return fmt.Errorf("%s evento de integraÃ§Ã£o: %w", acao, err)
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return errors.New("reserva do evento de integraÃ§Ã£o nÃ£o estÃ¡ mais ativa")
	}
	return nil
}
