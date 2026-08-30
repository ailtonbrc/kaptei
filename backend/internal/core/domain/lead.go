package domain

import "time"

type LeadStatus string

const (
	LeadStatusNovo          LeadStatus = "NOVO"
	LeadStatusEmAtendimento LeadStatus = "EM_ATENDIMENTO"
	LeadStatusQualificado   LeadStatus = "QUALIFICADO"
	LeadStatusDescartado    LeadStatus = "DESCARTADO"
)

type CapturaLeadWebhook struct {
	Nome     string `json:"nome"`
	Email    string `json:"email,omitempty"`
	Telefone string `json:"telefone,omitempty"`
	Origem   string `json:"origem,omitempty"`
	Mensagem string `json:"mensagem,omitempty"`
}

type CapturaLeadIntegracao struct {
	Nome              string
	ImovelID          *string
	Email             string
	Telefone          string
	Origem            string
	Mensagem          string
	ChaveIdempotencia string
}

type Lead struct {
	ID                  string     `json:"id"`
	ContaID             string     `json:"conta_id"`
	UsuarioID           *string    `json:"usuario_id"` // Se nulo, está na Caixa de Entrada Geral
	ImovelID            *string    `json:"imovel_id"`
	ClienteID           *string    `json:"cliente_id"`
	Nome                string     `json:"nome"`
	Email               *string    `json:"email"`
	Telefone            *string    `json:"telefone"`
	Origem              *string    `json:"origem"` // Site, ZAP, VivaReal, etc
	Mensagem            *string    `json:"mensagem"`
	Status              LeadStatus `json:"status"`
	MotivoDescarte      *string    `json:"motivo_descarte"`
	PaginaOrigem        *string    `json:"pagina_origem,omitempty"`
	UTMSource           *string    `json:"utm_source,omitempty"`
	UTMMedium           *string    `json:"utm_medium,omitempty"`
	UTMCampaign         *string    `json:"utm_campaign,omitempty"`
	ConsentimentoLGPD   bool       `json:"consentimento_lgpd"`
	ConsentimentoEm     *time.Time `json:"consentimento_lgpd_em,omitempty"`
	ConsentimentoVersao *string    `json:"consentimento_lgpd_versao,omitempty"`
	ChaveIdempotencia   *string    `json:"-"`
	CriadoEm            time.Time  `json:"criado_em"`
	AtualizadoEm        time.Time  `json:"atualizado_em"`
}
