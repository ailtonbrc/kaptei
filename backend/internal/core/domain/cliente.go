package domain

import "time"

type Cliente struct {
	ID             string               `json:"id"`
	ContaID        string               `json:"conta_id"`
	Nome           string               `json:"nome"`
	CPF            *string              `json:"cpf,omitempty"`
	DataNascimento *string              `json:"data_nascimento,omitempty"` // YYYY-MM-DD
	EstadoCivil    *string              `json:"estado_civil,omitempty"`
	Email          *string              `json:"email"`
	Telefone       *string              `json:"telefone"`
	StatusFunil    string               `json:"status_funil"` // NOVO, ATENDIMENTO, VISITA, PROPOSTA, FECHADO
	Origem         *string              `json:"origem"`
	InteresseTipo  *string              `json:"interesse_tipo"`
	Notas          *string              `json:"notas"`
	Tags           []string             `json:"tags"`
	Preferencias   *ClientePreferencias `json:"preferencias"`
	Financeiro     *ClienteFinanceiro   `json:"financeiro,omitempty"`
	OrigemUTM      *ClienteOrigemUTM    `json:"origem_utm,omitempty"`
	CorretorID     *string              `json:"corretor_id,omitempty"`
	Temperatura    *string              `json:"temperatura,omitempty"` // Quente, Morno, Frio
	ProximaAcao    *time.Time           `json:"proxima_acao,omitempty"`
	CriadoEm       time.Time            `json:"criado_em"`
	AtualizadoEm   time.Time            `json:"atualizado_em"`
}

type ClienteFinanceiro struct {
	RendaMensal          *float64 `json:"renda_mensal,omitempty"`
	PrecisaFinanciamento *string  `json:"precisa_financiamento,omitempty"` // Sim, Nao, JaAprovado
	PossuiFGTS           *bool    `json:"possui_fgts,omitempty"`
	FormaPagamento       *string  `json:"forma_pagamento,omitempty"` // AVista, Financiamento, Consorcio, Parcelado
}

type ClienteOrigemUTM struct {
	Canal          string `json:"canal,omitempty"`
	Campanha       string `json:"campanha,omitempty"`
	ImovelOrigemID string `json:"imovel_origem_id,omitempty"`
}

type ClientePreferencias struct {
	TipoImovel   []string `json:"tipo_imovel,omitempty"` // Apartamento, Casa, Lote, Comercial, Rural
	Finalidade   string   `json:"finalidade,omitempty"`  // Compra, Locacao
	OrcamentoMin *float64 `json:"orcamento_min,omitempty"`
	OrcamentoMax *float64 `json:"orcamento_max,omitempty"`
	Bairros      []string `json:"bairros,omitempty"`
	QuartosMin   *int     `json:"quartos_min,omitempty"`
}
