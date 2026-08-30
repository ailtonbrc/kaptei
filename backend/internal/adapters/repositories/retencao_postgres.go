package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type retencaoPostgres struct{ db *sql.DB }

func NewRetencaoPostgres(db *sql.DB) ports.RetencaoRepository { return &retencaoPostgres{db: db} }

func politicaPadrao(contaID string) *domain.PoliticaRetencao {
	return &domain.PoliticaRetencao{ContaID: contaID, DiasLeadsDescartados: 730, DiasClientesPerdidos: 1825, TamanhoLote: 200}
}

func (r *retencaoPostgres) ObterPolitica(ctx context.Context, contaID string) (*domain.PoliticaRetencao, error) {
	politica := politicaPadrao(contaID)
	err := r.db.QueryRowContext(ctx, `SELECT ativa,dias_leads_descartados,dias_clientes_perdidos,tamanho_lote,fundamento_legal,ultima_execucao_em,atualizado_em FROM politicas_retencao WHERE conta_id=$1`, contaID).Scan(
		&politica.Ativa, &politica.DiasLeadsDescartados, &politica.DiasClientesPerdidos, &politica.TamanhoLote, &politica.FundamentoLegal, &politica.UltimaExecucaoEm, &politica.AtualizadoEm,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return politica, nil
	}
	return politica, err
}

func (r *retencaoPostgres) SalvarPolitica(ctx context.Context, politica *domain.PoliticaRetencao, usuarioID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO politicas_retencao (conta_id,ativa,dias_leads_descartados,dias_clientes_perdidos,tamanho_lote,fundamento_legal,atualizado_por)
		VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (conta_id) DO UPDATE SET ativa=EXCLUDED.ativa,
		dias_leads_descartados=EXCLUDED.dias_leads_descartados,dias_clientes_perdidos=EXCLUDED.dias_clientes_perdidos,
		tamanho_lote=EXCLUDED.tamanho_lote,fundamento_legal=EXCLUDED.fundamento_legal,atualizado_por=EXCLUDED.atualizado_por,atualizado_em=now()`,
		politica.ContaID, politica.Ativa, politica.DiasLeadsDescartados, politica.DiasClientesPerdidos, politica.TamanhoLote, politica.FundamentoLegal, usuarioID)
	return err
}

const condicaoBloqueioLead = `NOT EXISTS (SELECT 1 FROM bloqueios_retencao b WHERE b.conta_id=l.conta_id AND b.tipo_recurso='LEAD' AND b.recurso_id=l.id AND (b.valido_ate IS NULL OR b.valido_ate>now()))`
const condicaoBloqueioCliente = `NOT EXISTS (SELECT 1 FROM bloqueios_retencao b WHERE b.conta_id=c.conta_id AND b.tipo_recurso='CLIENTE' AND b.recurso_id=c.id AND (b.valido_ate IS NULL OR b.valido_ate>now()))`

func (r *retencaoPostgres) GerarRelatorio(ctx context.Context, contaID string, politica *domain.PoliticaRetencao) (*domain.RelatorioRetencao, error) {
	relatorio := &domain.RelatorioRetencao{}
	limiteLead := time.Now().AddDate(0, 0, -politica.DiasLeadsDescartados)
	limiteCliente := time.Now().AddDate(0, 0, -politica.DiasClientesPerdidos)
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM leads l WHERE l.conta_id=$1 AND l.status='DESCARTADO' AND l.atualizado_em<$2 AND `+condicaoBloqueioLead, contaID, limiteLead).Scan(&relatorio.LeadsElegiveis); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM clientes c WHERE c.conta_id=$1 AND c.status_funil='PERDIDO' AND c.atualizado_em<$2 AND `+condicaoBloqueioCliente+`
		AND NOT EXISTS (SELECT 1 FROM leads l WHERE l.conta_id=c.conta_id AND l.cliente_id=c.id AND (l.status<>'DESCARTADO' OR l.atualizado_em>=$3))
		AND NOT EXISTS (SELECT 1 FROM agendamentos a WHERE a.conta_id=c.conta_id AND a.cliente_id=c.id AND a.status IN ('AGENDADO','CONFIRMADO') AND a.data_hora_fim>=now())`, contaID, limiteCliente, limiteLead).Scan(&relatorio.ClientesElegiveis); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM bloqueios_retencao WHERE conta_id=$1 AND (valido_ate IS NULL OR valido_ate>now())`, contaID).Scan(&relatorio.BloqueiosVigentes); err != nil {
		return nil, err
	}
	return relatorio, nil
}

func (r *retencaoPostgres) Executar(ctx context.Context, contaID, usuarioID string, politica *domain.PoliticaRetencao) (*domain.ResultadoRetencao, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	politicaBloqueada := &domain.PoliticaRetencao{ContaID: contaID}
	if err := tx.QueryRowContext(ctx, `SELECT ativa,dias_leads_descartados,dias_clientes_perdidos,tamanho_lote,fundamento_legal
		FROM politicas_retencao WHERE conta_id=$1 FOR UPDATE`, contaID).Scan(
		&politicaBloqueada.Ativa,
		&politicaBloqueada.DiasLeadsDescartados,
		&politicaBloqueada.DiasClientesPerdidos,
		&politicaBloqueada.TamanhoLote,
		&politicaBloqueada.FundamentoLegal,
	); err != nil {
		return nil, err
	}
	if !politicaBloqueada.Ativa || strings.TrimSpace(politicaBloqueada.FundamentoLegal) == "" {
		return nil, errors.New("política de retenção não está ativa e fundamentada")
	}
	limiteLead := time.Now().AddDate(0, 0, -politicaBloqueada.DiasLeadsDescartados)
	limiteCliente := time.Now().AddDate(0, 0, -politicaBloqueada.DiasClientesPerdidos)
	leads, err := selecionarIDs(ctx, tx, `SELECT l.id FROM leads l WHERE l.conta_id=$1 AND l.status='DESCARTADO' AND l.atualizado_em<$2 AND `+condicaoBloqueioLead+` ORDER BY l.atualizado_em,l.id FOR UPDATE SKIP LOCKED LIMIT $3`, contaID, limiteLead, politicaBloqueada.TamanhoLote)
	if err != nil {
		return nil, err
	}
	clientes, err := selecionarIDs(ctx, tx, `SELECT c.id FROM clientes c WHERE c.conta_id=$1 AND c.status_funil='PERDIDO' AND c.atualizado_em<$2 AND `+condicaoBloqueioCliente+`
		AND NOT EXISTS (SELECT 1 FROM leads l WHERE l.conta_id=c.conta_id AND l.cliente_id=c.id AND (l.status<>'DESCARTADO' OR l.atualizado_em>=$3))
		AND NOT EXISTS (SELECT 1 FROM agendamentos a WHERE a.conta_id=c.conta_id AND a.cliente_id=c.id AND a.status IN ('AGENDADO','CONFIRMADO') AND a.data_hora_fim>=now())
		ORDER BY c.atualizado_em,c.id FOR UPDATE SKIP LOCKED LIMIT $4`, contaID, limiteCliente, limiteLead, politicaBloqueada.TamanhoLote)
	if err != nil {
		return nil, err
	}
	if err := anonimizarLeadsRetencao(ctx, tx, contaID, leads); err != nil {
		return nil, err
	}
	if err := anonimizarClientesRetencao(ctx, tx, contaID, clientes); err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO execucoes_retencao (conta_id,usuario_id,leads_anonimizados,clientes_anonimizados,fundamento_legal) VALUES ($1,$2,$3,$4,$5)`, contaID, usuarioID, len(leads), len(clientes), politicaBloqueada.FundamentoLegal)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE politicas_retencao SET ultima_execucao_em=now() WHERE conta_id=$1`, contaID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	resultado := &domain.ResultadoRetencao{LeadsAnonimizados: len(leads), ClientesAnonimizados: len(clientes)}
	if restante, erroRelatorio := r.GerarRelatorio(ctx, contaID, politicaBloqueada); erroRelatorio == nil {
		resultado.RelatorioRestante = *restante
	}
	return resultado, nil
}

func selecionarIDs(ctx context.Context, tx *sql.Tx, query string, argumentos ...any) ([]string, error) {
	rows, err := tx.QueryContext(ctx, query, argumentos...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func marcadoresIDs(inicio, quantidade int) string {
	marcadores := make([]string, quantidade)
	for indice := range quantidade {
		marcadores[indice] = fmt.Sprintf("$%d", inicio+indice)
	}
	return strings.Join(marcadores, ",")
}

func argumentosIDs(contaID string, ids []string) []any {
	argumentos := make([]any, 1, len(ids)+1)
	argumentos[0] = contaID
	for _, id := range ids {
		argumentos = append(argumentos, id)
	}
	return argumentos
}

func anonimizarLeadsRetencao(ctx context.Context, tx *sql.Tx, contaID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	marcadores, argumentos := marcadoresIDs(2, len(ids)), argumentosIDs(contaID, ids)
	if _, err := tx.ExecContext(ctx, `DELETE FROM conversas_whatsapp WHERE conta_id=$1 AND lead_id IN (`+marcadores+`)`, argumentos...); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE leads SET nome='Titular anonimizado',email=NULL,telefone=NULL,mensagem=NULL,motivo_descarte=NULL,pagina_origem=NULL,utm_source=NULL,utm_medium=NULL,utm_campaign=NULL,consentimento_lgpd=FALSE,consentimento_lgpd_em=NULL,consentimento_lgpd_versao=NULL,atualizado_em=now() WHERE conta_id=$1 AND id IN (`+marcadores+`)`, argumentos...)
	return err
}

func anonimizarClientesRetencao(ctx context.Context, tx *sql.Tx, contaID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	marcadores, argumentos := marcadoresIDs(2, len(ids)), argumentosIDs(contaID, ids)
	if _, err := tx.ExecContext(ctx, `DELETE FROM conversas_whatsapp WHERE conta_id=$1 AND lead_id IN (SELECT id FROM leads WHERE conta_id=$1 AND cliente_id IN (`+marcadores+`))`, argumentos...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE leads SET nome='Titular anonimizado',email=NULL,telefone=NULL,mensagem=NULL,motivo_descarte=NULL,pagina_origem=NULL,utm_source=NULL,utm_medium=NULL,utm_campaign=NULL,consentimento_lgpd=FALSE,consentimento_lgpd_em=NULL,consentimento_lgpd_versao=NULL,atualizado_em=now() WHERE conta_id=$1 AND cliente_id IN (`+marcadores+`)`, argumentos...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agendamentos SET titulo='Atendimento anonimizado',descricao=NULL,updated_at=now() WHERE conta_id=$1 AND cliente_id IN (`+marcadores+`)`, argumentos...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE interacoes SET descricao='Interação anonimizada' WHERE conta_id=$1 AND cliente_id IN (`+marcadores+`)`, argumentos...); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE clientes SET nome='Titular anonimizado',email=NULL,telefone=NULL,notas=NULL,cpf=NULL,data_nascimento=NULL,estado_civil=NULL,financeiro='{}'::jsonb,origem_utm='{}'::jsonb,tags='[]'::jsonb,preferencias='{}'::jsonb,atualizado_em=now() WHERE conta_id=$1 AND id IN (`+marcadores+`)`, argumentos...)
	return err
}

func (r *retencaoPostgres) ListarBloqueios(ctx context.Context, contaID string) ([]domain.BloqueioRetencao, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,conta_id,tipo_recurso,recurso_id,motivo,valido_ate,criado_em FROM bloqueios_retencao WHERE conta_id=$1 ORDER BY criado_em DESC`, contaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dados := make([]domain.BloqueioRetencao, 0)
	for rows.Next() {
		var b domain.BloqueioRetencao
		if err := rows.Scan(&b.ID, &b.ContaID, &b.TipoRecurso, &b.RecursoID, &b.Motivo, &b.ValidoAte, &b.CriadoEm); err != nil {
			return nil, err
		}
		dados = append(dados, b)
	}
	return dados, rows.Err()
}

func (r *retencaoPostgres) SalvarBloqueio(ctx context.Context, bloqueio *domain.BloqueioRetencao, usuarioID string) error {
	tabela := "leads"
	if bloqueio.TipoRecurso == "CLIENTE" {
		tabela = "clientes"
	}
	var existe bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+tabela+` WHERE id=$1 AND conta_id=$2)`, bloqueio.RecursoID, bloqueio.ContaID).Scan(&existe); err != nil {
		return err
	}
	if !existe {
		return errors.New("recurso não encontrado nesta conta")
	}
	return r.db.QueryRowContext(ctx, `INSERT INTO bloqueios_retencao (conta_id,tipo_recurso,recurso_id,motivo,valido_ate,criado_por) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (conta_id,tipo_recurso,recurso_id) DO UPDATE SET motivo=EXCLUDED.motivo,valido_ate=EXCLUDED.valido_ate,criado_por=EXCLUDED.criado_por,criado_em=now()
		RETURNING id,criado_em`, bloqueio.ContaID, bloqueio.TipoRecurso, bloqueio.RecursoID, bloqueio.Motivo, bloqueio.ValidoAte, usuarioID).Scan(&bloqueio.ID, &bloqueio.CriadoEm)
}

func (r *retencaoPostgres) RemoverBloqueio(ctx context.Context, id, contaID string) error {
	resultado, err := r.db.ExecContext(ctx, `DELETE FROM bloqueios_retencao WHERE id=$1 AND conta_id=$2`, id, contaID)
	return exigirLinhaAlterada(resultado, err, "bloqueio de retenção não encontrado")
}
