package handlers

import (
	"net/http"
	"strings"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type DominioSiteHandler struct{ servico ports.DominioSiteService }

func NewDominioSiteHandler(servico ports.DominioSiteService) *DominioSiteHandler {
	return &DominioSiteHandler{servico: servico}
}

func (h *DominioSiteHandler) Obter(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioDominio(w, r)
	if !ok {
		return
	}
	dominio, err := h.servico.Obter(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	if dominio == nil {
		responderJSON(w, http.StatusOK, map[string]any{"configurado": false})
		return
	}
	responderJSON(w, http.StatusOK, dominio)
}

func (h *DominioSiteHandler) Configurar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioDominio(w, r)
	if !ok {
		return
	}
	var requisicao struct {
		Hostname string `json:"hostname"`
	}
	if err := decodificarJSONLimitado(w, r, &requisicao, 4*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "domínio inválido"})
		return
	}
	dominio, err := h.servico.Configurar(r.Context(), usuario.ContaID, usuario.Papel, requisicao.Hostname)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, dominio)
}

func (h *DominioSiteHandler) Verificar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioDominio(w, r)
	if !ok {
		return
	}
	dominio, err := h.servico.Verificar(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]any{"erro": err.Error(), "dominio": dominio})
		return
	}
	responderJSON(w, http.StatusOK, dominio)
}

func (h *DominioSiteHandler) ResolverPublico(w http.ResponseWriter, r *http.Request) {
	hostname := strings.TrimSpace(r.PathValue("hostname"))
	site, err := h.servico.ResolverPublico(r.Context(), hostname)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "não foi possível resolver o domínio"})
		return
	}
	if site == nil {
		responderJSON(w, http.StatusNotFound, map[string]string{"erro": "domínio não encontrado"})
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	responderJSON(w, http.StatusOK, site)
}

func usuarioDominio(w http.ResponseWriter, r *http.Request) (*domain.Usuario, bool) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario == nil {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return nil, false
	}
	return usuario, true
}
