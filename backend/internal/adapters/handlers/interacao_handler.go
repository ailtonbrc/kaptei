package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type InteracaoHandler struct {
	service ports.InteracaoService
}

func NewInteracaoHandler(service ports.InteracaoService) *InteracaoHandler {
	return &InteracaoHandler{
		service: service,
	}
}

func (h *InteracaoHandler) Create(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}
	clienteID := r.PathValue("cliente_id")

	var interacao domain.Interacao
	if err := decodificarJSONLimitado(w, r, &interacao, 32*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Dados invÃ¡lidos")
		return
	}

	interacao.ContaID = usuario.ContaID
	interacao.ClienteID = clienteID
	if err := h.service.Create(r.Context(), &interacao, usuario.ID, usuario.Papel); err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(interacao)
}

func (h *InteracaoHandler) ListByCliente(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}
	clienteID := r.PathValue("cliente_id")

	interacoes, err := h.service.ListByClienteID(r.Context(), clienteID, usuario.ContaID, usuario.ID, usuario.Papel)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(interacoes)
}

func (h *InteracaoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}
	id := r.PathValue("id")

	if err := h.service.Delete(r.Context(), id, usuario.ContaID, usuario.ID, usuario.Papel); err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
