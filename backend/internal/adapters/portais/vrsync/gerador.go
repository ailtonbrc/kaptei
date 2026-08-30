package vrsync

import (
	"encoding/xml"
	"fmt"
	"math"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
)

const (
	esquemaVRSync = "http://www.vivareal.com/schemas/1.0/VRSync"
	localEsquema  = "http://xml.vivareal.com/vrsync.xsd"
)

var possivelHTML = regexp.MustCompile(`<[/!A-Za-z][^>]*>`)
var somenteDigitos = regexp.MustCompile(`\D`)

type Gerador struct{}

func NovoGerador() *Gerador { return &Gerador{} }

func (g *Gerador) Validar(dados *domain.DadosFeedPortal) *domain.DiagnosticoFeedPortal {
	diagnostico := &domain.DiagnosticoFeedPortal{ErrosGerais: make([]string, 0), Publicacoes: make([]domain.PublicacaoPortal, 0, len(dados.Anuncios))}
	if !dados.Configuracao.Ativa {
		diagnostico.ErrosGerais = append(diagnostico.ErrosGerais, "a integração está desativada")
	}
	if strings.TrimSpace(dados.Configuracao.NomeContato) == "" {
		diagnostico.ErrosGerais = append(diagnostico.ErrosGerais, "nome de contato é obrigatório")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(dados.Configuracao.EmailContato)); err != nil {
		diagnostico.ErrosGerais = append(diagnostico.ErrosGerais, "e-mail de contato é inválido")
	}
	if !dados.SitePublicado || strings.TrimSpace(dados.SiteSlug) == "" {
		diagnostico.ErrosGerais = append(diagnostico.ErrosGerais, "o site público precisa estar publicado")
	}

	for _, anuncio := range dados.Anuncios {
		erros := validarAnuncio(anuncio)
		publicacao := domain.PublicacaoPortal{ImovelID: anuncio.ID, Titulo: anuncio.Titulo, Tipo: anuncio.Tipo, Finalidade: anuncio.Finalidade, Status: anuncio.Status, Ativa: true, TipoPublicacao: anuncio.TipoPublicacao, Erros: erros}
		diagnostico.Publicacoes = append(diagnostico.Publicacoes, publicacao)
		diagnostico.TotalSelecionado++
		if len(erros) == 0 {
			diagnostico.TotalValido++
		}
	}
	sort.Slice(diagnostico.Publicacoes, func(i, j int) bool { return diagnostico.Publicacoes[i].Titulo < diagnostico.Publicacoes[j].Titulo })
	diagnostico.Valido = len(diagnostico.ErrosGerais) == 0 && diagnostico.TotalSelecionado > 0 && diagnostico.TotalSelecionado == diagnostico.TotalValido
	if diagnostico.TotalSelecionado == 0 {
		diagnostico.ErrosGerais = append(diagnostico.ErrosGerais, "nenhum imóvel foi selecionado para o portal")
	}
	return diagnostico
}

func (g *Gerador) Gerar(dados *domain.DadosFeedPortal, origemPublica string, instante time.Time) ([]byte, error) {
	diagnostico := g.Validar(dados)
	if !diagnostico.Valido {
		return nil, fmt.Errorf("feed VRSync inválido: %s", resumirDiagnostico(diagnostico))
	}
	origem, err := url.Parse(strings.TrimRight(origemPublica, "/"))
	if err != nil || origem.Scheme != "https" || origem.Host == "" {
		return nil, fmt.Errorf("origem pública HTTPS inválida")
	}

	feed := listingDataFeed{
		XMLNS:          esquemaVRSync,
		XMLNSXSI:       "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: esquemaVRSync + " " + localEsquema,
		Header:         header{Provider: "Kaptei", Email: dados.Configuracao.EmailContato, ContactName: dados.Configuracao.NomeContato, PublishDate: instante.UTC().Format(time.RFC3339), Telephone: dados.Configuracao.TelefoneContato},
		Listings:       make([]listing, 0, len(dados.Anuncios)),
	}
	for _, anuncio := range dados.Anuncios {
		feed.Listings = append(feed.Listings, montarListing(anuncio, dados, origem))
	}
	conteudo, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("codificar feed VRSync: %w", err)
	}
	return append([]byte(xml.Header), conteudo...), nil
}

func validarAnuncio(anuncio domain.AnuncioPortal) []string {
	erros := make([]string, 0)
	if anuncio.Status != "Ativo" {
		erros = append(erros, "imóvel não está ativo")
	}
	if tamanho := utf8.RuneCountInString(strings.TrimSpace(anuncio.Titulo)); tamanho < 10 || tamanho > 100 {
		erros = append(erros, "título deve ter entre 10 e 100 caracteres")
	}
	descricao := ""
	if anuncio.Descricao != nil {
		descricao = strings.TrimSpace(*anuncio.Descricao)
	}
	if tamanho := utf8.RuneCountInString(descricao); tamanho < 50 || tamanho > 3000 {
		erros = append(erros, "descrição deve ter entre 50 e 3.000 caracteres")
	}
	if _, ok := tipoPropriedade(anuncio.Tipo); !ok {
		erros = append(erros, "tipo de imóvel não possui mapeamento VRSync")
		if possivelHTML.MatchString(anuncio.Titulo) || possivelHTML.MatchString(descricao) {
			erros = append(erros, "título e descrição não podem conter tags HTML")
		}
	}
	if _, ok := tipoTransacao(anuncio.Finalidade); !ok {
		erros = append(erros, "finalidade não possui mapeamento VRSync")
	}
	if anuncio.Finalidade == "Venda" || anuncio.Finalidade == "Venda e Locação" {
		if anuncio.ValorVenda == nil || *anuncio.ValorVenda <= 0 {
			erros = append(erros, "valor de venda é obrigatório")
		}
	}
	if anuncio.Finalidade == "Locação" || anuncio.Finalidade == "Venda e Locação" {
		if anuncio.ValorLocacao == nil || *anuncio.ValorLocacao <= 0 {
			erros = append(erros, "valor de locação é obrigatório")
		}
	}
	if anuncio.CEP == nil || len(somenteDigitos.ReplaceAllString(*anuncio.CEP, "")) != 8 {
		erros = append(erros, "CEP válido é obrigatório")
	}
	if anuncio.Cidade == nil || strings.TrimSpace(*anuncio.Cidade) == "" {
		erros = append(erros, "cidade é obrigatória")
	}
	if anuncio.Bairro == nil || strings.TrimSpace(*anuncio.Bairro) == "" {
		erros = append(erros, "bairro é obrigatório")
	}
	if anuncio.Estado == nil || nomeEstado(strings.ToUpper(strings.TrimSpace(*anuncio.Estado))) == "" {
		erros = append(erros, "UF válida é obrigatória")
	}
	if exigeAreaTotal(anuncio.Tipo) {
		if anuncio.AreaTotal == nil || *anuncio.AreaTotal <= 0 {
			erros = append(erros, "área total é obrigatória para este tipo")
			if strings.TrimSpace(anuncio.Slug) == "" {
				erros = append(erros, "URL pública do imóvel é obrigatória")
			}
		}
	} else if anuncio.AreaUtil == nil || *anuncio.AreaUtil <= 0 {
		erros = append(erros, "área útil é obrigatória para este tipo")
	}
	if len(anuncio.Fotos) == 0 {
		erros = append(erros, "ao menos uma foto JPG é obrigatória")
	}
	capa := 0
	for _, foto := range anuncio.Fotos {
		if foto.IsCapa {
			capa++
		}
		if foto.TipoConteudo == nil || *foto.TipoConteudo != "image/jpeg" {
			erros = append(erros, "todas as fotos devem estar em formato JPG")
			break
		}
		if foto.TamanhoBytes == nil || *foto.TamanhoBytes > 7*1024*1024 {
			erros = append(erros, "cada foto deve possuir no máximo 7 MB")
			break
		}
		endereco, err := url.Parse(foto.URL)
		if err != nil || (endereco.Scheme != "https" && endereco.Scheme != "http") || endereco.Host == "" {
			erros = append(erros, "todas as fotos precisam de URL pública válida")
			break
		}
	}
	if capa > 1 {
		erros = append(erros, "somente uma foto pode ser definida como capa")
	}
	if anuncio.TipoPublicacao != "STANDARD" && anuncio.TipoPublicacao != "PREMIUM" && anuncio.TipoPublicacao != "SUPER_PREMIUM" {
		erros = append(erros, "tipo de publicação inválido")
	}
	return erros
}

func resumirDiagnostico(diagnostico *domain.DiagnosticoFeedPortal) string {
	partes := append([]string{}, diagnostico.ErrosGerais...)
	for _, publicacao := range diagnostico.Publicacoes {
		if len(publicacao.Erros) > 0 {
			partes = append(partes, publicacao.ImovelID+": "+strings.Join(publicacao.Erros, ", "))
		}
	}
	return strings.Join(partes, "; ")
}

func tipoTransacao(finalidade string) (string, bool) {
	valor, ok := map[string]string{"Venda": "For Sale", "Locação": "For Rent", "Venda e Locação": "Sale/Rent"}[finalidade]
	return valor, ok
}

func tipoPropriedade(tipo string) (string, bool) {
	valor, ok := map[string]string{"Casa": "Residential / Home", "Apartamento": "Residential / Apartment", "Terreno": "Residential / Land Lot", "Comercial": "Commercial / Building", "Galpão": "Commercial / Industrial", "Rural": "Residential / Agricultural"}[tipo]
	return valor, ok
}

func exigeAreaTotal(tipo string) bool {
	return tipo == "Terreno" || tipo == "Galpão" || tipo == "Rural"
}

func nomeEstado(uf string) string {
	return map[string]string{"AC": "Acre", "AL": "Alagoas", "AP": "Amapá", "AM": "Amazonas", "BA": "Bahia", "CE": "Ceará", "DF": "Distrito Federal", "ES": "Espírito Santo", "GO": "Goiás", "MA": "Maranhão", "MT": "Mato Grosso", "MS": "Mato Grosso do Sul", "MG": "Minas Gerais", "PA": "Pará", "PB": "Paraíba", "PR": "Paraná", "PE": "Pernambuco", "PI": "Piauí", "RJ": "Rio de Janeiro", "RN": "Rio Grande do Norte", "RS": "Rio Grande do Sul", "RO": "Rondônia", "RR": "Roraima", "SC": "Santa Catarina", "SP": "São Paulo", "SE": "Sergipe", "TO": "Tocantins"}[uf]
}

type listingDataFeed struct {
	XMLName        xml.Name  `xml:"ListingDataFeed"`
	XMLNS          string    `xml:"xmlns,attr"`
	XMLNSXSI       string    `xml:"xmlns:xsi,attr"`
	SchemaLocation string    `xml:"xsi:schemaLocation,attr"`
	Header         header    `xml:"Header"`
	Listings       []listing `xml:"Listings>Listing"`
}

type header struct {
	Provider    string `xml:"Provider"`
	Email       string `xml:"Email"`
	ContactName string `xml:"ContactName"`
	PublishDate string `xml:"PublishDate"`
	Telephone   string `xml:"Telephone,omitempty"`
}
type listing struct {
	ListingID       string      `xml:"ListingID"`
	Title           string      `xml:"Title"`
	TransactionType string      `xml:"TransactionType"`
	PublicationType string      `xml:"PublicationType"`
	DetailViewURL   string      `xml:"DetailViewUrl"`
	Media           media       `xml:"Media"`
	Details         details     `xml:"Details"`
	Location        location    `xml:"Location"`
	ContactInfo     contactInfo `xml:"ContactInfo"`
}
type media struct {
	Items []mediaItem `xml:"Item"`
}
type mediaItem struct {
	Medium  string `xml:"medium,attr"`
	Caption string `xml:"caption,attr,omitempty"`
	Primary *bool  `xml:"primary,attr,omitempty"`
	URL     string `xml:",chardata"`
}
type dinheiro struct {
	Currency string `xml:"currency,attr"`
	Period   string `xml:"period,attr,omitempty"`
	Valor    int64  `xml:",chardata"`
}
type area struct {
	Unit  string `xml:"unit,attr"`
	Valor int64  `xml:",chardata"`
}
type garagem struct {
	Type  string `xml:"type,attr"`
	Valor int    `xml:",chardata"`
}
type details struct {
	UsageType                 string    `xml:"UsageType"`
	PropertyType              string    `xml:"PropertyType"`
	Description               string    `xml:"Description,cdata"`
	ListPrice                 *dinheiro `xml:"ListPrice,omitempty"`
	RentalPrice               *dinheiro `xml:"RentalPrice,omitempty"`
	PropertyAdministrationFee *dinheiro `xml:"PropertyAdministrationFee,omitempty"`
	Iptu                      *dinheiro `xml:"Iptu,omitempty"`
	LotArea                   *area     `xml:"LotArea,omitempty"`
	LivingArea                *area     `xml:"LivingArea,omitempty"`
	Bedrooms                  int       `xml:"Bedrooms"`
	Bathrooms                 int       `xml:"Bathrooms"`
	Suites                    int       `xml:"Suites,omitempty"`
	Garage                    *garagem  `xml:"Garage,omitempty"`
}
type textoAbreviado struct {
	Abbreviation string `xml:"abbreviation,attr"`
	Valor        string `xml:",chardata"`
}
type location struct {
	DisplayAddress string         `xml:"displayAddress,attr"`
	Country        textoAbreviado `xml:"Country"`
	State          textoAbreviado `xml:"State"`
	City           string         `xml:"City"`
	Neighborhood   string         `xml:"Neighborhood"`
	Address        string         `xml:"Address,omitempty"`
	StreetNumber   string         `xml:"StreetNumber,omitempty"`
	Complement     string         `xml:"Complement,omitempty"`
	PostalCode     string         `xml:"PostalCode"`
}
type contactInfo struct {
	Name      string `xml:"Name"`
	Email     string `xml:"Email"`
	Website   string `xml:"Website,omitempty"`
	Telephone string `xml:"Telephone,omitempty"`
}

func montarListing(anuncio domain.AnuncioPortal, dados *domain.DadosFeedPortal, origem *url.URL) listing {
	transacao, _ := tipoTransacao(anuncio.Finalidade)
	propriedade, _ := tipoPropriedade(anuncio.Tipo)
	uso := "Residential"
	if strings.HasPrefix(propriedade, "Commercial") {
		uso = "Commercial"
	}
	detalhes := details{UsageType: uso, PropertyType: propriedade, Description: strings.TrimSpace(*anuncio.Descricao), Bedrooms: anuncio.Quartos, Bathrooms: anuncio.Banheiros}
	if anuncio.Suites > 0 {
		detalhes.Suites = anuncio.Suites
	}
	if anuncio.Vagas > 0 {
		detalhes.Garage = &garagem{Type: "Parking Space", Valor: anuncio.Vagas}
	}
	if anuncio.ValorVenda != nil {
		detalhes.ListPrice = &dinheiro{Currency: "BRL", Valor: int64(math.Trunc(*anuncio.ValorVenda))}
	}
	if anuncio.ValorLocacao != nil {
		detalhes.RentalPrice = &dinheiro{Currency: "BRL", Period: "Monthly", Valor: int64(math.Trunc(*anuncio.ValorLocacao))}
	}
	if anuncio.ValorCondominio != nil && *anuncio.ValorCondominio > 0 {
		detalhes.PropertyAdministrationFee = &dinheiro{Currency: "BRL", Valor: int64(math.Trunc(*anuncio.ValorCondominio))}
	}
	if anuncio.ValorIPTU != nil && *anuncio.ValorIPTU > 0 {
		detalhes.Iptu = &dinheiro{Currency: "BRL", Period: "Yearly", Valor: int64(math.Trunc(*anuncio.ValorIPTU))}
	}
	if anuncio.AreaTotal != nil && *anuncio.AreaTotal > 0 {
		detalhes.LotArea = &area{Unit: "square metres", Valor: int64(math.Trunc(*anuncio.AreaTotal))}
	}
	if anuncio.AreaUtil != nil && *anuncio.AreaUtil > 0 {
		detalhes.LivingArea = &area{Unit: "square metres", Valor: int64(math.Trunc(*anuncio.AreaUtil))}
	}
	itens := make([]mediaItem, 0, len(anuncio.Fotos))
	for indice, foto := range anuncio.Fotos {
		primaria := foto.IsCapa || (indice == 0 && !temCapa(anuncio.Fotos))
		var p *bool
		if primaria {
			valor := true
			p = &valor
		}
		itens = append(itens, mediaItem{Medium: "image", Caption: fmt.Sprintf("foto-%d", indice+1), Primary: p, URL: foto.URL})
	}
	uf := strings.ToUpper(strings.TrimSpace(*anuncio.Estado))
	display := map[string]string{"BAIRRO": "Neighborhood", "LOGRADOURO": "Street", "COMPLETO": "All"}[dados.Configuracao.ExibicaoEndereco]
	local := location{DisplayAddress: display, Country: textoAbreviado{Abbreviation: "BR", Valor: "Brasil"}, State: textoAbreviado{Abbreviation: uf, Valor: nomeEstado(uf)}, City: valorTexto(anuncio.Cidade), Neighborhood: valorTexto(anuncio.Bairro), Address: valorTexto(anuncio.Logradouro), StreetNumber: valorTexto(anuncio.Numero), Complement: valorTexto(anuncio.Complemento), PostalCode: valorTexto(anuncio.CEP)}
	detalheURL := *origem
	detalheURL.Path = strings.TrimRight(detalheURL.Path, "/") + "/s/" + url.PathEscape(dados.SiteSlug) + "/imoveis/" + url.PathEscape(anuncio.Slug)
	return listing{ListingID: anuncio.ID, Title: strings.TrimSpace(anuncio.Titulo), TransactionType: transacao, PublicationType: anuncio.TipoPublicacao, DetailViewURL: detalheURL.String(), Media: media{Items: itens}, Details: detalhes, Location: local, ContactInfo: contactInfo{Name: dados.Configuracao.NomeContato, Email: dados.Configuracao.EmailContato, Website: origem.String() + "/s/" + url.PathEscape(dados.SiteSlug), Telephone: dados.Configuracao.TelefoneContato}}
}

func temCapa(fotos []domain.ImovelFoto) bool {
	for _, foto := range fotos {
		if foto.IsCapa {
			return true
		}
	}
	return false
}
func valorTexto(valor *string) string {
	if valor == nil {
		return ""
	}
	return strings.TrimSpace(*valor)
}
