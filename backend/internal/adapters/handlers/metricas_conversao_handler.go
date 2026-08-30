package handlers

import (
	"errors"
	"net/http"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"github.com/msdev/kaptei/internal/core/services"
)

type MetricasConversaoHandler struct {
	servico ports.MetricasConversaoService
}

func NewMetricasConversaoHandler(servico ports.MetricasConversaoService) *MetricasConversaoHandler {
	return &MetricasConversaoHandler{servico: servico}
}

func (h *MetricasConversaoHandler) Registrar(w http.ResponseWriter, r *http.Request) {
	var evento domain.EventoConversaoPublico
	if err := decodificarJSONLimitado(w, r, &evento, 8*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "evento inválido"})
		return
	}
	if err := h.servico.Registrar(r.Context(), r.PathValue("slug"), evento); err != nil {
		switch {
		case errors.Is(err, services.ErrEventoConversaoInvalido):
			responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "evento inválido"})
		case errors.Is(err, services.ErrSiteConversaoNaoEncontrado):
			responderJSON(w, http.StatusNotFound, map[string]string{"erro": "site não encontrado"})
		default:
			responderErro(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusAccepted)
}
