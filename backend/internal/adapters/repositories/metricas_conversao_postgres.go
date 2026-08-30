package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type metricasConversaoPostgres struct {
	banco *sql.DB
}

func NewMetricasConversaoPostgres(banco *sql.DB) ports.MetricasConversaoRepository {
	return &metricasConversaoPostgres{banco: banco}
}

func (r *metricasConversaoPostgres) Registrar(ctx context.Context, evento *domain.EventoConversao) error {
	_, err := r.banco.ExecContext(ctx, `
		INSERT INTO eventos_conversao_site (
			conta_id, chave_evento, sessao_id, tipo, imovel_id,
			utm_source, utm_medium, utm_campaign
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT DO NOTHING`,
		evento.ContaID, evento.ChaveEvento, evento.SessaoID, evento.Tipo, evento.ImovelID,
		evento.UTMSource, evento.UTMMedium, evento.UTMCampaign,
	)
	if err != nil {
		return fmt.Errorf("registrar evento de conversão: %w", err)
	}
	return nil
}

func (r *metricasConversaoPostgres) ObterResumo(ctx context.Context, contaID string, desde time.Time) (*domain.ResumoConversaoSite, error) {
	resumo := &domain.ResumoConversaoSite{Fontes: make(map[string]int)}
	err := r.banco.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT sessao_id) FILTER (WHERE tipo='SITE_VISUALIZADO'),
			COUNT(DISTINCT sessao_id) FILTER (WHERE tipo='IMOVEL_VISUALIZADO'),
			COUNT(DISTINCT sessao_id) FILTER (WHERE tipo='FORMULARIO_INICIADO'),
			COUNT(DISTINCT sessao_id) FILTER (WHERE tipo='LEAD_ENVIADO'),
			COUNT(DISTINCT sessao_id) FILTER (WHERE tipo='WHATSAPP_CLICADO'),
			COUNT(DISTINCT sessao_id) FILTER (WHERE tipo='TELEFONE_CLICADO')
		FROM eventos_conversao_site
		WHERE conta_id=$1 AND criado_em >= $2 AND expira_em > NOW()`, contaID, desde,
	).Scan(
		&resumo.VisitasSite, &resumo.ImoveisVisualizados, &resumo.FormulariosIniciados,
		&resumo.ContatosEnviados, &resumo.CliquesWhatsApp, &resumo.CliquesTelefone,
	)
	if err != nil {
		return nil, fmt.Errorf("resumir eventos de conversão: %w", err)
	}
	if resumo.VisitasSite > 0 {
		resumo.TaxaConversao = float64(resumo.ContatosEnviados) * 100 / float64(resumo.VisitasSite)
	}

	linhas, err := r.banco.QueryContext(ctx, `
		SELECT COALESCE(NULLIF(utm_source,''), 'Orgânico'), COUNT(DISTINCT sessao_id)
		FROM eventos_conversao_site
		WHERE conta_id=$1 AND tipo='SITE_VISUALIZADO' AND criado_em >= $2 AND expira_em > NOW()
		GROUP BY COALESCE(NULLIF(utm_source,''), 'Orgânico')
		ORDER BY COUNT(DISTINCT sessao_id) DESC`, contaID, desde)
	if err != nil {
		return nil, fmt.Errorf("agrupar fontes de conversão: %w", err)
	}
	defer linhas.Close()
	for linhas.Next() {
		var fonte string
		var total int
		if err := linhas.Scan(&fonte, &total); err != nil {
			return nil, fmt.Errorf("ler fonte de conversão: %w", err)
		}
		resumo.Fontes[fonte] = total
	}
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("percorrer fontes de conversão: %w", err)
	}
	return resumo, nil
}

func (r *metricasConversaoPostgres) ExpurgarExpirados(ctx context.Context) (int64, error) {
	resultado, err := r.banco.ExecContext(ctx, `DELETE FROM eventos_conversao_site WHERE expira_em <= NOW()`)
	if err != nil {
		return 0, fmt.Errorf("expurgar métricas de conversão expiradas: %w", err)
	}
	total, err := resultado.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("contar métricas de conversão expurgadas: %w", err)
	}
	return total, nil
}
