package handlers

import (
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type RetencaoHandler struct{ servico ports.RetencaoService }

func NewRetencaoHandler(servico ports.RetencaoService) *RetencaoHandler {
	return &RetencaoHandler{servico: servico}
}

func (h *RetencaoHandler) ObterPolitica(w http.ResponseWriter, r *http.Request) {
	u, ok := usuarioRetencao(w, r)
	if !ok {
		return
	}
	politica, err := h.servico.ObterPolitica(r.Context(), u.ContaID, u.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, politica)
}
func (h *RetencaoHandler) SalvarPolitica(w http.ResponseWriter, r *http.Request) {
	u, ok := usuarioRetencao(w, r)
	if !ok {
		return
	}
	var politica domain.PoliticaRetencao
	if decodificarJSONLimitado(w, r, &politica, 12*1024) != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "política inválida"})
		return
	}
	if err := h.servico.SalvarPolitica(r.Context(), u.ContaID, u.ID, u.Papel, politica); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"status": "salva"})
}
func (h *RetencaoHandler) Relatorio(w http.ResponseWriter, r *http.Request) {
	u, ok := usuarioRetencao(w, r)
	if !ok {
		return
	}
	relatorio, err := h.servico.GerarRelatorio(r.Context(), u.ContaID, u.Papel)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, relatorio)
}
func (h *RetencaoHandler) Executar(w http.ResponseWriter, r *http.Request) {
	u, ok := usuarioRetencao(w, r)
	if !ok {
		return
	}
	var req struct {
		Confirmacao string `json:"confirmacao"`
	}
	if decodificarJSONLimitado(w, r, &req, 2*1024) != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "confirmação inválida"})
		return
	}
	resultado, err := h.servico.Executar(r.Context(), u.ContaID, u.ID, u.Papel, req.Confirmacao)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, resultado)
}
func (h *RetencaoHandler) ListarBloqueios(w http.ResponseWriter, r *http.Request) {
	u, ok := usuarioRetencao(w, r)
	if !ok {
		return
	}
	dados, err := h.servico.ListarBloqueios(r.Context(), u.ContaID, u.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, dados)
}
func (h *RetencaoHandler) SalvarBloqueio(w http.ResponseWriter, r *http.Request) {
	u, ok := usuarioRetencao(w, r)
	if !ok {
		return
	}
	var b domain.BloqueioRetencao
	if decodificarJSONLimitado(w, r, &b, 8*1024) != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "bloqueio inválido"})
		return
	}
	salvo, err := h.servico.SalvarBloqueio(r.Context(), u.ContaID, u.ID, u.Papel, b)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusCreated, salvo)
}
func (h *RetencaoHandler) RemoverBloqueio(w http.ResponseWriter, r *http.Request) {
	u, ok := usuarioRetencao(w, r)
	if !ok {
		return
	}
	if err := h.servico.RemoverBloqueio(r.Context(), r.PathValue("id"), u.ContaID, u.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func usuarioRetencao(w http.ResponseWriter, r *http.Request) (*domain.Usuario, bool) {
	u, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || u == nil {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return nil, false
	}
	return u, true
}
