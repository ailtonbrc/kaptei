package domain

import "time"

type RotaSitemap struct {
	SlugSite     string
	SlugImovel   *string
	AtualizadoEm time.Time
}

type ConfiguracaoSitePublico struct {
	LogoURL     string `json:"logo_url,omitempty"`
	CorPrimaria string `json:"cor_primaria,omitempty"`
	Titulo      string `json:"titulo,omitempty"`
	Subtitulo   string `json:"subtitulo,omitempty"`
	Descricao   string `json:"descricao,omitempty"`
	Telefone    string `json:"telefone,omitempty"`
	WhatsApp    string `json:"whatsapp,omitempty"`
	Email       string `json:"email,omitempty"`
	Cidade      string `json:"cidade,omitempty"`
	CRECI       string `json:"creci,omitempty"`
}

type SitePublico struct {
	ContaID      string                  `json:"-"`
	Slug         string                  `json:"slug"`
	Nome         string                  `json:"nome"`
	Publicado    bool                    `json:"publicado"`
	Configuracao ConfiguracaoSitePublico `json:"configuracao"`
}

type FiltrosCatalogoPublico struct {
	Tipo       string
	Finalidade string
	Cidade     string
	Bairro     string
	ValorMin   *float64
	ValorMax   *float64
	QuartosMin *int
	Pagina     int
	Limite     int
}

type ImovelPublico struct {
	ID              string       `json:"id"`
	Slug            string       `json:"slug"`
	Titulo          string       `json:"titulo"`
	Tipo            string       `json:"tipo"`
	Finalidade      string       `json:"finalidade"`
	ValorVenda      *float64     `json:"valor_venda,omitempty"`
	ValorLocacao    *float64     `json:"valor_locacao,omitempty"`
	ValorCondominio *float64     `json:"valor_condominio,omitempty"`
	ValorIPTU       *float64     `json:"valor_iptu,omitempty"`
	AreaTotal       *float64     `json:"area_total,omitempty"`
	AreaUtil        *float64     `json:"area_util,omitempty"`
	Quartos         int          `json:"quartos"`
	Suites          int          `json:"suites"`
	Banheiros       int          `json:"banheiros"`
	Vagas           int          `json:"vagas"`
	Bairro          *string      `json:"bairro,omitempty"`
	Cidade          *string      `json:"cidade,omitempty"`
	Estado          *string      `json:"estado,omitempty"`
	Descricao       *string      `json:"descricao,omitempty"`
	TituloSEO       *string      `json:"titulo_seo,omitempty"`
	DescricaoSEO    *string      `json:"descricao_seo,omitempty"`
	Destaque        bool         `json:"destaque"`
	Fotos           []ImovelFoto `json:"fotos"`
}

type CapturaLeadPublico struct {
	Nome              string  `json:"nome"`
	Email             *string `json:"email,omitempty"`
	Telefone          *string `json:"telefone,omitempty"`
	Mensagem          *string `json:"mensagem,omitempty"`
	ImovelSlug        *string `json:"imovel_slug,omitempty"`
	PaginaOrigem      *string `json:"pagina_origem,omitempty"`
	UTMSource         *string `json:"utm_source,omitempty"`
	UTMMedium         *string `json:"utm_medium,omitempty"`
	UTMCampaign       *string `json:"utm_campaign,omitempty"`
	ConsentimentoLGPD bool    `json:"consentimento_lgpd"`
	ChaveIdempotencia string  `json:"chave_idempotencia"`
	Website           string  `json:"website,omitempty"`
}
