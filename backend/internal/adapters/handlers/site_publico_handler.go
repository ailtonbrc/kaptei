package handlers

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type SitePublicoHandler struct {
	servico    ports.SitePublicoService
	urlPublica string
}

func NewSitePublicoHandler(servico ports.SitePublicoService, urlPublica string) *SitePublicoHandler {
	return &SitePublicoHandler{servico: servico, urlPublica: strings.TrimRight(urlPublica, "/")}
}

type urlSitemap struct {
	Loc               string `xml:"loc"`
	UltimaModificacao string `xml:"lastmod,omitempty"`
}

type conjuntoURLs struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []urlSitemap `xml:"url"`
}

func (h *SitePublicoHandler) Sitemap(w http.ResponseWriter, r *http.Request) {
	rotas, err := h.servico.ListarRotasSitemap(r.Context())
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "não foi possível gerar o sitemap"})
		return
	}
	base := h.urlPublica
	if base == "" {
		base = "https://" + r.Host
	}
	urls := make([]urlSitemap, 0, len(rotas)*2)
	sitesIncluidos := make(map[string]bool)
	for _, rota := range rotas {
		if !sitesIncluidos[rota.SlugSite] {
			urls = append(urls, urlSitemap{Loc: base + "/s/" + url.PathEscape(rota.SlugSite), UltimaModificacao: rota.AtualizadoEm.UTC().Format("2006-01-02")})
			sitesIncluidos[rota.SlugSite] = true
		}
		if rota.SlugImovel != nil {
			urls = append(urls, urlSitemap{Loc: base + "/s/" + url.PathEscape(rota.SlugSite) + "/imoveis/" + url.PathEscape(*rota.SlugImovel), UltimaModificacao: rota.AtualizadoEm.UTC().Format("2006-01-02")})
		}
	}
	dados, err := xml.Marshal(conjuntoURLs{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls})
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "não foi possível gerar o sitemap"})
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=900")
	_, _ = w.Write(append([]byte(xml.Header), dados...))
}

func (h *SitePublicoHandler) Robots(w http.ResponseWriter, r *http.Request) {
	base := h.urlPublica
	if base == "" {
		base = "https://" + r.Host
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte("User-agent: *\nAllow: /s/\nDisallow: /app/\nSitemap: " + base + "/sitemap.xml\n"))
}

func (h *SitePublicoHandler) GetPublico(w http.ResponseWriter, r *http.Request) {
	site, err := h.servico.GetPublico(r.Context(), r.PathValue("slug"))
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "não foi possível carregar o site"})
		return
	}
	if site == nil {
		responderJSON(w, http.StatusNotFound, map[string]string{"erro": "site não encontrado"})
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	responderJSON(w, http.StatusOK, site)
}

func (h *SitePublicoHandler) ListarImoveis(w http.ResponseWriter, r *http.Request) {
	filtros, err := lerFiltrosCatalogo(r)
	if err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": err.Error()})
		return
	}
	imoveis, total, err := h.servico.ListarImoveis(r.Context(), r.PathValue("slug"), filtros)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "não foi possível carregar os imóveis"})
		return
	}
	if imoveis == nil {
		responderJSON(w, http.StatusNotFound, map[string]string{"erro": "site não encontrado"})
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=30, stale-while-revalidate=120")
	responderJSON(w, http.StatusOK, map[string]any{
		"dados": imoveis, "total": total, "pagina": filtros.Pagina, "limite": filtros.Limite,
	})
}

func (h *SitePublicoHandler) GetImovel(w http.ResponseWriter, r *http.Request) {
	imovel, err := h.servico.GetImovel(r.Context(), r.PathValue("slug"), r.PathValue("slug_imovel"))
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "não foi possível carregar o imóvel"})
		return
	}
	if imovel == nil {
		responderJSON(w, http.StatusNotFound, map[string]string{"erro": "imóvel não encontrado"})
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=30, stale-while-revalidate=120")
	responderJSON(w, http.StatusOK, imovel)
}

func (h *SitePublicoHandler) CapturarLead(w http.ResponseWriter, r *http.Request) {
	var captura domain.CapturaLeadPublico
	if err := decodificarJSONLimitado(w, r, &captura, 32*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados de contato inválidos"})
		return
	}
	if err := h.servico.CapturarLead(r.Context(), r.PathValue("slug"), captura); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusCreated, map[string]string{"mensagem": "contato recebido com sucesso"})
}

func (h *SitePublicoHandler) GetAdministracao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioPodeAdministrarSite(r)
	if !ok {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": "sem permissão para administrar o site"})
		return
	}
	site, err := h.servico.GetAdministracao(r.Context(), usuario.ContaID)
	if err != nil || site == nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "não foi possível carregar a configuração"})
		return
	}
	responderJSON(w, http.StatusOK, site)
}

func (h *SitePublicoHandler) Salvar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioPodeAdministrarSite(r)
	if !ok {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": "sem permissão para administrar o site"})
		return
	}
	var site domain.SitePublico
	if err := decodificarJSONLimitado(w, r, &site, 72*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "configuração inválida"})
		return
	}
	site.ContaID = usuario.ContaID
	if err := h.servico.Salvar(r.Context(), &site); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "site atualizado com sucesso"})
}

func usuarioPodeAdministrarSite(r *http.Request) (*domain.Usuario, bool) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario == nil {
		return nil, false
	}
	permitido := usuario.Papel == domain.RoleSuperAdmin || usuario.Papel == domain.RoleGestor || usuario.Papel == domain.RoleCorretorSolo
	return usuario, permitido
}

func decodificarJSONLimitado(w http.ResponseWriter, r *http.Request, destino any, limite int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, limite)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destino); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("corpo JSON deve conter um único objeto")
	}
	return nil
}

func lerFiltrosCatalogo(r *http.Request) (domain.FiltrosCatalogoPublico, error) {
	query := r.URL.Query()
	filtros := domain.FiltrosCatalogoPublico{
		Tipo: strings.TrimSpace(query.Get("tipo")), Finalidade: strings.TrimSpace(query.Get("finalidade")),
		Cidade: strings.TrimSpace(query.Get("cidade")), Bairro: strings.TrimSpace(query.Get("bairro")),
	}
	for _, valor := range []string{filtros.Tipo, filtros.Finalidade, filtros.Cidade, filtros.Bairro} {
		if utf8.RuneCountInString(valor) > 120 {
			return filtros, errors.New("filtro de texto excede o limite permitido")
		}
	}
	var err error
	if filtros.Pagina, err = inteiroOpcional(query.Get("pagina"), 1); err != nil {
		return filtros, errors.New("página inválida")
	}
	if filtros.Pagina < 1 || filtros.Pagina > 100_000 {
		return filtros, errors.New("página fora do intervalo permitido")
	}
	if filtros.Limite, err = inteiroOpcional(query.Get("limite"), 12); err != nil {
		return filtros, errors.New("limite inválido")
	}
	if filtros.Limite < 1 || filtros.Limite > 48 {
		return filtros, errors.New("limite deve estar entre 1 e 48")
	}
	if valor := query.Get("quartos_min"); valor != "" {
		numero, e := strconv.Atoi(valor)
		if e != nil || numero < 0 || numero > 100 {
			return filtros, errors.New("quantidade de quartos inválida")
		}
		filtros.QuartosMin = &numero
	}
	if filtros.ValorMin, err = decimalOpcional(query.Get("valor_min")); err != nil {
		return filtros, errors.New("valor mínimo inválido")
	}
	if filtros.ValorMax, err = decimalOpcional(query.Get("valor_max")); err != nil {
		return filtros, errors.New("valor máximo inválido")
	}
	if filtros.ValorMin != nil && filtros.ValorMax != nil && *filtros.ValorMin > *filtros.ValorMax {
		return filtros, errors.New("valor mínimo não pode superar o valor máximo")
	}
	return filtros, nil
}

func inteiroOpcional(valor string, padrao int) (int, error) {
	if valor == "" {
		return padrao, nil
	}
	return strconv.Atoi(valor)
}
func decimalOpcional(valor string) (*float64, error) {
	if valor == "" {
		return nil, nil
	}
	numero, err := strconv.ParseFloat(valor, 64)
	if err != nil || math.IsNaN(numero) || math.IsInf(numero, 0) || numero < 0 || numero > 1_000_000_000_000 {
		return nil, errors.New("decimal inválido")
	}
	return &numero, nil
}
