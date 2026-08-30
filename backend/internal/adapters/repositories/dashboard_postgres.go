package repositories

import (
	"context"
	"database/sql"
	"github.com/msdev/kaptei/internal/core/ports"
)

type DashboardPostgres struct {
	db *sql.DB
}

func NewDashboardPostgres(db *sql.DB) ports.DashboardRepository {
	return &DashboardPostgres{db: db}
}

func (r *DashboardPostgres) GetFunilConversao(ctx context.Context, contaID string, usuarioID *string) (map[string]int, error) {
	query := `
		SELECT status_funil, COUNT(*) 
		FROM clientes 
		WHERE conta_id = $1 AND ($2::uuid IS NULL OR corretor_id=$2)
		GROUP BY status_funil
	`
	rows, err := r.db.QueryContext(ctx, query, contaID, usuarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	funil := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		funil[status] = count
	}
	return funil, nil
}

func (r *DashboardPostgres) GetOrigemLeads(ctx context.Context, contaID string, usuarioID *string) (map[string]int, error) {
	query := `
		SELECT origem, COUNT(*) 
		FROM leads 
		WHERE conta_id = $1 AND ($2::uuid IS NULL OR usuario_id=$2)
		GROUP BY origem
	`
	rows, err := r.db.QueryContext(ctx, query, contaID, usuarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	origens := make(map[string]int)
	for rows.Next() {
		var origem sql.NullString
		var count int
		if err := rows.Scan(&origem, &count); err != nil {
			return nil, err
		}
		if origem.Valid && origem.String != "" {
			origens[origem.String] = count
		} else {
			origens["ORGÂNICO"] = origens["ORGÂNICO"] + count
		}
	}
	return origens, nil
}

func (r *DashboardPostgres) GetMetricasResumo(ctx context.Context, contaID string, usuarioID *string) (map[string]interface{}, error) {
	var totalImoveis, totalClientes, leadsRecentes, visitasPendentes int
	err := r.db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM imoveis WHERE conta_id=$1 AND status='Ativo'),
		(SELECT COUNT(*) FROM clientes WHERE conta_id=$1 AND ($2::uuid IS NULL OR corretor_id=$2)),
		(SELECT COUNT(*) FROM leads WHERE conta_id=$1 AND ($2::uuid IS NULL OR usuario_id=$2) AND criado_em >= NOW() - INTERVAL '30 days'),
		(SELECT COUNT(*) FROM agendamentos WHERE conta_id=$1 AND ($2::uuid IS NULL OR usuario_id=$2) AND tipo='VISITA' AND data_hora_inicio >= NOW() AND status IN ('AGENDADO','CONFIRMADO'))`, contaID, usuarioID).
		Scan(&totalImoveis, &totalClientes, &leadsRecentes, &visitasPendentes)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_imoveis": totalImoveis, "total_clientes": totalClientes,
		"leads_30_dias": leadsRecentes, "visitas_pendentes": visitasPendentes,
	}, nil
}

func (r *DashboardPostgres) GetEvolucaoLeads(ctx context.Context, contaID string, usuarioID *string) ([]string, []int, error) {
	rows, err := r.db.QueryContext(ctx, `WITH meses AS (
		SELECT generate_series(date_trunc('month', NOW()) - INTERVAL '5 months', date_trunc('month', NOW()), INTERVAL '1 month') AS mes
	) SELECT TO_CHAR(m.mes, 'YYYY-MM'), COUNT(l.id)
	FROM meses m LEFT JOIN leads l ON l.conta_id=$1 AND ($2::uuid IS NULL OR l.usuario_id=$2) AND date_trunc('month', l.criado_em)=m.mes
	GROUP BY m.mes ORDER BY m.mes`, contaID, usuarioID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	categorias, valores := make([]string, 0, 6), make([]int, 0, 6)
	for rows.Next() {
		var categoria string
		var valor int
		if err := rows.Scan(&categoria, &valor); err != nil {
			return nil, nil, err
		}
		categorias = append(categorias, categoria)
		valores = append(valores, valor)
	}
	return categorias, valores, rows.Err()
}
