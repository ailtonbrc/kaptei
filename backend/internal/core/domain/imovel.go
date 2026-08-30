package domain

import (
	"time"
)

type Imovel struct {
	ID              string    `json:"id"`
	ContaID         string    `json:"conta_id"`
	UsuarioID       string    `json:"usuario_id"`
	Titulo          string    `json:"titulo"`
	Tipo            string    `json:"tipo"`
	Finalidade      string    `json:"finalidade"`
	Status          string    `json:"status"`
	ValorVenda      *float64  `json:"valor_venda"`
	ValorLocacao    *float64  `json:"valor_locacao"`
	ValorCondominio *float64  `json:"valor_condominio"`
	ValorIPTU       *float64  `json:"valor_iptu"`
	AreaTotal       *float64  `json:"area_total"`
	AreaUtil        *float64  `json:"area_util"`
	Quartos         int       `json:"quartos"`
	Suites          int       `json:"suites"`
	Banheiros       int       `json:"banheiros"`
	Vagas           int       `json:"vagas"`
	CEP             *string   `json:"cep"`
	Logradouro      *string   `json:"logradouro"`
	Numero          *string   `json:"numero"`
	Complemento     *string   `json:"complemento"`
	Bairro          *string   `json:"bairro"`
	Cidade          *string   `json:"cidade"`
	Estado          *string   `json:"estado"`
	Descricao       *string   `json:"descricao"`
	SlugPublico     *string   `json:"slug_publico,omitempty"`
	Publicado       bool      `json:"publicado"`
	Destaque        bool      `json:"destaque"`
	TituloSEO       *string   `json:"titulo_seo,omitempty"`
	DescricaoSEO    *string   `json:"descricao_seo,omitempty"`
	CriadoEm        time.Time `json:"criado_em"`
	AtualizadoEm    time.Time `json:"atualizado_em"`

	Fotos []ImovelFoto `json:"fotos,omitempty"`
}

type ImovelFoto struct {
	ID              string    `json:"id"`
	ImovelID        string    `json:"imovel_id"`
	URL             string    `json:"url"`
	URLThumbnail    *string   `json:"url_thumbnail,omitempty"`
	ChaveObjeto     *string   `json:"-"`
	ChaveThumbnail  *string   `json:"-"`
	ProvedorStorage *string   `json:"-"`
	TipoConteudo    *string   `json:"tipo_conteudo,omitempty"`
	TamanhoBytes    *int64    `json:"tamanho_bytes,omitempty"`
	Largura         *int      `json:"largura,omitempty"`
	Altura          *int      `json:"altura,omitempty"`
	HashSHA256      *string   `json:"hash_sha256,omitempty"`
	Ordem           int       `json:"ordem"`
	IsCapa          bool      `json:"is_capa"`
	CriadoEm        time.Time `json:"criado_em"`
}

type ImagemProcessada struct {
	Principal    []byte
	Thumbnail    []byte
	TipoConteudo string
	Extensao     string
	Largura      int
	Altura       int
	HashSHA256   string
}
