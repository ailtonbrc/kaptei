package domain

import "time"

type TipoEventoConversao string

const (
	EventoSiteVisualizado    TipoEventoConversao = "SITE_VISUALIZADO"
	EventoImovelVisualizado  TipoEventoConversao = "IMOVEL_VISUALIZADO"
	EventoFormularioIniciado TipoEventoConversao = "FORMULARIO_INICIADO"
	EventoLeadEnviado        TipoEventoConversao = "LEAD_ENVIADO"
	EventoWhatsAppClicado    TipoEventoConversao = "WHATSAPP_CLICADO"
	EventoTelefoneClicado    TipoEventoConversao = "TELEFONE_CLICADO"
)

type EventoConversaoPublico struct {
	ChaveEvento string              `json:"chave_evento"`
	SessaoID    string              `json:"sessao_id"`
	Tipo        TipoEventoConversao `json:"tipo"`
	ImovelSlug  *string             `json:"imovel_slug,omitempty"`
	UTMSource   *string             `json:"utm_source,omitempty"`
	UTMMedium   *string             `json:"utm_medium,omitempty"`
	UTMCampaign *string             `json:"utm_campaign,omitempty"`
}

type EventoConversao struct {
	EventoConversaoPublico
	ContaID  string
	ImovelID *string
	CriadoEm time.Time
}

type ResumoConversaoSite struct {
	VisitasSite          int            `json:"visitas_site"`
	ImoveisVisualizados  int            `json:"imoveis_visualizados"`
	FormulariosIniciados int            `json:"formularios_iniciados"`
	ContatosEnviados     int            `json:"contatos_enviados"`
	CliquesWhatsApp      int            `json:"cliques_whatsapp"`
	CliquesTelefone      int            `json:"cliques_telefone"`
	TaxaConversao        float64        `json:"taxa_conversao"`
	Fontes               map[string]int `json:"fontes"`
}
