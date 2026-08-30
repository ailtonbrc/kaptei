package domain

import "time"

const (
	ProvedorMeta             = "META"
	TipoEventoMetaLeadGerado = "LEAD_GERADO"
)

type ConfiguracaoMetaLeads struct {
	ID                     string    `json:"id,omitempty"`
	ContaID                string    `json:"-"`
	PaginaID               string    `json:"pagina_id"`
	TokenPaginaProtegido   string    `json:"-"`
	TokenPaginaConfigurado bool      `json:"token_pagina_configurado"`
	DisponivelNoServidor   bool      `json:"disponivel_no_servidor"`
	Ativa                  bool      `json:"ativa"`
	CriadoEm               time.Time `json:"criado_em,omitempty"`
	AtualizadoEm           time.Time `json:"atualizado_em,omitempty"`
}

type AtualizacaoMetaLeads struct {
	PaginaID    string `json:"pagina_id"`
	TokenPagina string `json:"token_pagina,omitempty"`
	Ativa       bool   `json:"ativa"`
}

type EventoIntegracao struct {
	ID                   string
	ContaID              string
	Provedor             string
	Tipo                 string
	IdentificadorExterno string
	PaginaID             string
	FormularioID         *string
	AnuncioID            *string
	OcorridoEm           *time.Time
	Status               string
	Tentativas           int
	MaximoTentativas     int
	DisponivelEm         time.Time
	BloqueadoAte         *time.Time
	BloqueadoPor         *string
	CriadoEm             time.Time
	PayloadProtegido     string
}

type LeadMeta struct {
	ID           string
	CriadoEm     *time.Time
	AnuncioID    *string
	FormularioID *string
	Campos       map[string][]string
}
