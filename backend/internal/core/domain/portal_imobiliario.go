package domain

import "time"

const PortalGrupoOLX = "GRUPO_OLX"

type ConfiguracaoPortal struct {
	ID               string    `json:"id,omitempty"`
	ContaID          string    `json:"-"`
	Portal           string    `json:"portal"`
	Ativa            bool      `json:"ativa"`
	TokenFeedPrefixo *string   `json:"token_feed_prefixo,omitempty"`
	NomeContato      string    `json:"nome_contato"`
	EmailContato     string    `json:"email_contato"`
	TelefoneContato  string    `json:"telefone_contato"`
	ExibicaoEndereco string    `json:"exibicao_endereco"`
	AtualizadoEm     time.Time `json:"atualizado_em,omitempty"`
}

type CredencialFeedPortal struct {
	Token      string `json:"token"`
	URLFeed    string `json:"url_feed"`
	URLWebhook string `json:"url_webhook"`
}

type PublicacaoPortal struct {
	ImovelID       string   `json:"imovel_id"`
	Titulo         string   `json:"titulo"`
	Tipo           string   `json:"tipo"`
	Finalidade     string   `json:"finalidade"`
	Status         string   `json:"status"`
	Ativa          bool     `json:"ativa"`
	TipoPublicacao string   `json:"tipo_publicacao"`
	Erros          []string `json:"erros"`
}

type AtualizacaoPublicacaoPortal struct {
	Ativa          bool   `json:"ativa"`
	TipoPublicacao string `json:"tipo_publicacao"`
}

type AnuncioPortal struct {
	ID              string
	Slug            string
	Titulo          string
	Tipo            string
	Finalidade      string
	Status          string
	ValorVenda      *float64
	ValorLocacao    *float64
	ValorCondominio *float64
	ValorIPTU       *float64
	AreaTotal       *float64
	AreaUtil        *float64
	Quartos         int
	Suites          int
	Banheiros       int
	Vagas           int
	CEP             *string
	Logradouro      *string
	Numero          *string
	Complemento     *string
	Bairro          *string
	Cidade          *string
	Estado          *string
	Descricao       *string
	TipoPublicacao  string
	Fotos           []ImovelFoto
}

type DadosFeedPortal struct {
	Configuracao  ConfiguracaoPortal
	NomeConta     string
	SiteSlug      string
	SitePublicado bool
	Anuncios      []AnuncioPortal
}

type DiagnosticoFeedPortal struct {
	Valido           bool               `json:"valido"`
	TotalSelecionado int                `json:"total_selecionado"`
	TotalValido      int                `json:"total_valido"`
	ErrosGerais      []string           `json:"erros_gerais"`
	Publicacoes      []PublicacaoPortal `json:"publicacoes"`
}
