package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type preRenderizador struct {
	indice     string
	apiBase    string
	urlPublica string
	cliente    *http.Client
}

type siteSEO struct {
	Slug         string `json:"slug"`
	Nome         string `json:"nome"`
	Configuracao struct {
		LogoURL   string `json:"logo_url"`
		Titulo    string `json:"titulo"`
		Subtitulo string `json:"subtitulo"`
		Descricao string `json:"descricao"`
		Telefone  string `json:"telefone"`
		Email     string `json:"email"`
		Cidade    string `json:"cidade"`
		CRECI     string `json:"creci"`
	} `json:"configuracao"`
}

type fotoSEO struct {
	URL string `json:"url"`
}
type imovelSEO struct {
	Slug         string    `json:"slug"`
	Titulo       string    `json:"titulo"`
	Tipo         string    `json:"tipo"`
	Finalidade   string    `json:"finalidade"`
	ValorVenda   *float64  `json:"valor_venda"`
	ValorLocacao *float64  `json:"valor_locacao"`
	Quartos      int       `json:"quartos"`
	Banheiros    int       `json:"banheiros"`
	Vagas        int       `json:"vagas"`
	Cidade       *string   `json:"cidade"`
	Bairro       *string   `json:"bairro"`
	Estado       *string   `json:"estado"`
	Descricao    *string   `json:"descricao"`
	TituloSEO    *string   `json:"titulo_seo"`
	DescricaoSEO *string   `json:"descricao_seo"`
	Fotos        []fotoSEO `json:"fotos"`
}

type paginaImoveisSEO struct {
	Dados []imovelSEO `json:"dados"`
}

func novoPreRenderizador(diretorio, apiBase, urlPublica string) (*preRenderizador, error) {
	indice, err := os.ReadFile(filepath.Join(diretorio, "index.html"))
	if err != nil {
		return nil, fmt.Errorf("ler index para pré-renderização: %w", err)
	}
	apiBase = strings.TrimRight(strings.TrimSpace(apiBase), "/")
	if endereco, err := url.ParseRequestURI(apiBase); err != nil || !endereco.IsAbs() || (endereco.Scheme != "http" && endereco.Scheme != "https") {
		return nil, fmt.Errorf("endereço da API para pré-renderização é inválido")
	}
	return &preRenderizador{indice: string(indice), apiBase: apiBase, urlPublica: strings.TrimRight(strings.TrimSpace(urlPublica), "/"), cliente: &http.Client{Timeout: 4 * time.Second}}, nil
}

func (p *preRenderizador) Tentar(w http.ResponseWriter, r *http.Request) bool {
	partes := partesRotaPublica(r.URL.Path)
	dominioPersonalizado := partes == nil
	if p.tentarArquivoSEODominio(w, r) {
		return true
	}
	ctx, cancelar := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancelar()
	var site siteSEO
	if partes != nil {
		if p.obterJSON(ctx, "/public/sites/"+url.PathEscape(partes.slug), &site) != nil {
			return false
		}
	} else {
		partes = partesRotaDominio(r.URL.Path)
		hostname, valido := hostnameRequisicao(r.Host)
		if partes == nil || !valido || p.obterJSON(ctx, "/public/dominios/"+url.PathEscape(hostname), &site) != nil {
			return false
		}
		partes.slug = site.Slug
	}
	canonical := p.canonical(r, r.URL.Path)
	if partes.privacidade {
		titulo := "Política de privacidade | " + site.Nome
		descricao := "Como " + site.Nome + " trata dados pessoais e atende aos direitos dos titulares."
		p.responder(w, titulo, descricao, canonical, site.Configuracao.LogoURL, dadosEstruturadosSite(site, canonical), corpoPrivacidade(site))
		return true
	}
	if partes.imovel != "" {
		var imovel imovelSEO
		if p.obterJSON(ctx, "/public/sites/"+url.PathEscape(partes.slug)+"/imoveis/"+url.PathEscape(partes.imovel), &imovel) != nil {
			return false
		}
		titulo := imovel.Titulo + " | " + site.Nome
		if imovel.TituloSEO != nil && strings.TrimSpace(*imovel.TituloSEO) != "" {
			titulo = *imovel.TituloSEO + " | " + site.Nome
		}
		descricao := descricaoImovel(imovel, site)
		imagem := ""
		if len(imovel.Fotos) > 0 {
			imagem = imovel.Fotos[0].URL
		}
		p.responder(w, titulo, descricao, canonical, imagem, dadosEstruturadosImovel(site, imovel, canonical, imagem), ajustarLinksDominio(corpoImovel(site, imovel), site.Slug, dominioPersonalizado))
		return true
	}
	var pagina paginaImoveisSEO
	_ = p.obterJSON(ctx, "/public/sites/"+url.PathEscape(partes.slug)+"/imoveis?pagina=1&limite=6", &pagina)
	titulo := site.Configuracao.Titulo
	if titulo == "" {
		titulo = "Imóveis selecionados por " + site.Nome
	}
	descricao := site.Configuracao.Descricao
	if descricao == "" {
		descricao = site.Configuracao.Subtitulo
	}
	if descricao == "" {
		descricao = "Encontre imóveis para comprar ou alugar com " + site.Nome + "."
	}
	p.responder(w, titulo+" | "+site.Nome, descricao, canonical, site.Configuracao.LogoURL, dadosEstruturadosSite(site, canonical), ajustarLinksDominio(corpoSite(site, pagina.Dados), site.Slug, dominioPersonalizado))
	return true
}

type rotaPublica struct {
	slug, imovel string
	privacidade  bool
}

func partesRotaPublica(caminho string) *rotaPublica {
	partes := strings.Split(strings.Trim(caminho, "/"), "/")
	if len(partes) < 2 || partes[0] != "s" || partes[1] == "" {
		return nil
	}
	rota := &rotaPublica{slug: partes[1]}
	if len(partes) == 2 {
		return rota
	}
	if len(partes) == 3 && partes[2] == "privacidade" {
		rota.privacidade = true
		return rota
	}
	if len(partes) == 4 && partes[2] == "imoveis" && partes[3] != "" {
		rota.imovel = partes[3]
		return rota
	}
	return nil
}

func (p *preRenderizador) obterJSON(ctx context.Context, caminho string, destino any) error {
	requisicao, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiBase+caminho, nil)
	if err != nil {
		return err
	}
	resposta, err := p.cliente.Do(requisicao)
	if err != nil {
		return err
	}
	defer resposta.Body.Close()
	if resposta.StatusCode != http.StatusOK {
		return fmt.Errorf("API respondeu %d", resposta.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(resposta.Body, 2<<20))
	return decoder.Decode(destino)
}

func (p *preRenderizador) canonical(r *http.Request, caminho string) string {
	base := p.urlPublica
	if !strings.HasPrefix(caminho, "/s/") {
		hostname, valido := hostnameRequisicao(r.Host)
		if valido {
			esquema := "https"
			if hostname == "localhost" || hostname == "127.0.0.1" {
				esquema = "http"
			}
			return esquema + "://" + r.Host + caminho
		}
	}
	if base == "" {
		esquema := "https"
		if strings.HasPrefix(r.Host, "localhost") || strings.HasPrefix(r.Host, "127.0.0.1") {
			esquema = "http"
		}
		base = esquema + "://" + r.Host
	}
	return base + caminho
}

func (p *preRenderizador) responder(w http.ResponseWriter, titulo, descricao, canonical, imagem string, estruturados any, corpo template.HTML) {
	metadados := struct {
		Titulo, Descricao, Canonical, Imagem string
		JSONLD                               template.JS
	}{Titulo: titulo, Descricao: descricao, Canonical: canonical, Imagem: imagem}
	jsonLD, _ := json.Marshal(estruturados)
	metadados.JSONLD = template.JS(jsonLD)
	const modeloHead = `<title>{{.Titulo}}</title><meta name="description" content="{{.Descricao}}"><link rel="canonical" href="{{.Canonical}}"><meta property="og:type" content="website"><meta property="og:title" content="{{.Titulo}}"><meta property="og:description" content="{{.Descricao}}"><meta property="og:url" content="{{.Canonical}}">{{if .Imagem}}<meta property="og:image" content="{{.Imagem}}"><meta name="twitter:card" content="summary_large_image">{{end}}<script type="application/ld+json">{{.JSONLD}}</script>`
	var head strings.Builder
	_ = template.Must(template.New("head").Parse(modeloHead)).Execute(&head, metadados)
	html := strings.Replace(p.indice, "<title>Kaptei - Portal de Vendas</title>", "", 1)
	html = strings.Replace(html, "</head>", head.String()+"</head>", 1)
	html = strings.Replace(html, `<div id="root"></div>`, `<div id="root">`+string(corpo)+`</div>`, 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

func dadosEstruturadosSite(site siteSEO, canonical string) map[string]any {
	return map[string]any{"@context": "https://schema.org", "@type": "RealEstateAgent", "name": site.Nome, "url": canonical, "telephone": site.Configuracao.Telefone, "email": site.Configuracao.Email, "areaServed": site.Configuracao.Cidade}
}

func dadosEstruturadosImovel(site siteSEO, imovel imovelSEO, canonical, imagem string) map[string]any {
	return map[string]any{"@context": "https://schema.org", "@type": "RealEstateListing", "name": imovel.Titulo, "url": canonical, "image": imagem, "description": descricaoImovel(imovel, site), "offeredBy": dadosEstruturadosSite(site, canonical)}
}

func descricaoImovel(imovel imovelSEO, site siteSEO) string {
	if imovel.DescricaoSEO != nil && strings.TrimSpace(*imovel.DescricaoSEO) != "" {
		return *imovel.DescricaoSEO
	}
	if imovel.Descricao != nil && strings.TrimSpace(*imovel.Descricao) != "" {
		texto := strings.TrimSpace(*imovel.Descricao)
		if len([]rune(texto)) > 300 {
			return string([]rune(texto)[:300])
		}
		return texto
	}
	return imovel.Titulo + " anunciado por " + site.Nome + "."
}

func renderizarCorpo(dados any, modelo string) template.HTML {
	var saida strings.Builder
	_ = template.Must(template.New("corpo").Funcs(template.FuncMap{"valor": func(v *float64) string {
		if v == nil {
			return "Consulte"
		}
		return fmt.Sprintf("R$ %.2f", *v)
	}}).Parse(modelo)).Execute(&saida, dados)
	return template.HTML(saida.String())
}

func corpoSite(site siteSEO, imoveis []imovelSEO) template.HTML {
	dados := struct {
		Site    siteSEO
		Imoveis []imovelSEO
	}{site, imoveis}
	return renderizarCorpo(dados, `<main style="font-family:system-ui;max-width:1120px;margin:auto;padding:48px 20px"><header><h1>{{.Site.Configuracao.Titulo}}</h1><p>{{.Site.Configuracao.Subtitulo}}</p></header><section><h2>Imóveis selecionados</h2>{{range .Imoveis}}<article><h3><a href="/s/{{$.Site.Slug}}/imoveis/{{.Slug}}">{{.Titulo}}</a></h3><p>{{.Tipo}} para {{.Finalidade}} · {{valor .ValorVenda}}</p></article>{{end}}</section><footer><p>{{.Site.Nome}} {{.Site.Configuracao.CRECI}}</p></footer></main>`)
}

func corpoImovel(site siteSEO, imovel imovelSEO) template.HTML {
	dados := struct {
		Site   siteSEO
		Imovel imovelSEO
	}{site, imovel}
	return renderizarCorpo(dados, `<main style="font-family:system-ui;max-width:960px;margin:auto;padding:48px 20px"><a href="/s/{{.Site.Slug}}">Voltar aos imóveis</a><h1>{{.Imovel.Titulo}}</h1><p>{{.Imovel.Tipo}} para {{.Imovel.Finalidade}}</p>{{if .Imovel.Fotos}}<img src="{{(index .Imovel.Fotos 0).URL}}" alt="{{.Imovel.Titulo}}" style="max-width:100%;height:auto">{{end}}<h2>Características</h2><p>{{.Imovel.Quartos}} quartos · {{.Imovel.Banheiros}} banheiros · {{.Imovel.Vagas}} vagas</p><p>{{.Imovel.Descricao}}</p><strong>{{valor .Imovel.ValorVenda}}</strong></main>`)
}

func corpoPrivacidade(site siteSEO) template.HTML {
	return renderizarCorpo(site, `<main style="font-family:system-ui;max-width:760px;margin:auto;padding:48px 20px"><h1>Política de privacidade</h1><p>Controlador dos dados: {{.Nome}}</p><h2>Dados e finalidade</h2><p>Os dados informados nos formulários são utilizados para responder ao contato e apresentar oportunidades imobiliárias relacionadas.</p><h2>Seus direitos</h2><p>Você pode solicitar confirmação, acesso, correção, anonimização, bloqueio, exclusão, portabilidade, informação sobre compartilhamento ou revogação do consentimento.</p><p>Canal do controlador: {{.Configuracao.Email}}</p></main>`)
}
