package main

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type urlSitemapDominio struct {
	Loc string `xml:"loc"`
}
type conjuntoSitemapDominio struct {
	XMLName xml.Name            `xml:"urlset"`
	Xmlns   string              `xml:"xmlns,attr"`
	URLs    []urlSitemapDominio `xml:"url"`
}
type paginaDominioSEO struct {
	Dados  []imovelSEO `json:"dados"`
	Total  int         `json:"total"`
	Pagina int         `json:"pagina"`
	Limite int         `json:"limite"`
}

func (p *preRenderizador) tentarArquivoSEODominio(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/robots.txt" && r.URL.Path != "/sitemap.xml" {
		return false
	}
	hostname, valido := hostnameRequisicao(r.Host)
	if !valido {
		return false
	}
	ctx, cancelar := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancelar()
	var site siteSEO
	if p.obterJSON(ctx, "/public/dominios/"+url.PathEscape(hostname), &site) != nil {
		return false
	}
	base := baseDoHost(r, hostname)
	if r.URL.Path == "/robots.txt" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write([]byte("User-agent: *\nAllow: /\nDisallow: /app/\nSitemap: " + base + "/sitemap.xml\n"))
		return true
	}
	urls := []urlSitemapDominio{{Loc: base + "/"}, {Loc: base + "/privacidade"}}
	for pagina := 1; pagina <= 1000; pagina++ {
		var resultado paginaDominioSEO
		caminho := "/public/sites/" + url.PathEscape(site.Slug) + "/imoveis?limite=48&pagina=" + strconv.Itoa(pagina)
		if p.obterJSON(ctx, caminho, &resultado) != nil {
			return false
		}
		for _, imovel := range resultado.Dados {
			urls = append(urls, urlSitemapDominio{Loc: base + "/imoveis/" + url.PathEscape(imovel.Slug)})
		}
		if pagina*48 >= resultado.Total {
			break
		}
	}
	dados, err := xml.Marshal(conjuntoSitemapDominio{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls})
	if err != nil {
		return false
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=900")
	_, _ = w.Write(append([]byte(xml.Header), dados...))
	return true
}

func baseDoHost(r *http.Request, hostname string) string {
	esquema := "https"
	if hostname == "localhost" || hostname == "127.0.0.1" {
		esquema = "http"
	}
	return esquema + "://" + strings.TrimSuffix(r.Host, ".")
}
