package domain

import "time"

type PoliticaRetencao struct {
	ContaID              string     `json:"-"`
	Ativa                bool       `json:"ativa"`
	DiasLeadsDescartados int        `json:"dias_leads_descartados"`
	DiasClientesPerdidos int        `json:"dias_clientes_perdidos"`
	TamanhoLote          int        `json:"tamanho_lote"`
	FundamentoLegal      string     `json:"fundamento_legal"`
	UltimaExecucaoEm     *time.Time `json:"ultima_execucao_em,omitempty"`
	AtualizadoEm         *time.Time `json:"atualizado_em,omitempty"`
}

type RelatorioRetencao struct {
	LeadsElegiveis    int `json:"leads_elegiveis"`
	ClientesElegiveis int `json:"clientes_elegiveis"`
	BloqueiosVigentes int `json:"bloqueios_vigentes"`
}

type ResultadoRetencao struct {
	LeadsAnonimizados    int               `json:"leads_anonimizados"`
	ClientesAnonimizados int               `json:"clientes_anonimizados"`
	RelatorioRestante    RelatorioRetencao `json:"relatorio_restante"`
}

type BloqueioRetencao struct {
	ID          string     `json:"id"`
	ContaID     string     `json:"-"`
	TipoRecurso string     `json:"tipo_recurso"`
	RecursoID   string     `json:"recurso_id"`
	Motivo      string     `json:"motivo"`
	ValidoAte   *time.Time `json:"valido_ate,omitempty"`
	CriadoEm    time.Time  `json:"criado_em"`
}
