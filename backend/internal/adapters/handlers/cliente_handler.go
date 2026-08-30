package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type ClienteHandler struct {
	service ports.ClienteService
}

func NewClienteHandler(s ports.ClienteService) *ClienteHandler {
	return &ClienteHandler{service: s}
}

func (h *ClienteHandler) Create(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	var req domain.Cliente
	if err := decodificarJSONLimitado(w, r, &req, 64*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Dados invÃ¡lidos")
		return
	}

	req.ContaID = usuario.ContaID

	if err := h.service.Create(r.Context(), &req, usuario.ID, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func (h *ClienteHandler) List(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	clientes, err := h.service.List(r.Context(), usuario.ContaID, usuario.ID, usuario.Papel, filtroPaginacao(r))
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientes)
}

func (h *ClienteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		responderErro(w, http.StatusBadRequest, "ID obrigatÃ³rio")
		return
	}

	cliente, err := h.service.GetByID(r.Context(), id, usuario.ContaID, usuario.ID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	if cliente == nil {
		responderErro(w, http.StatusNotFound, "Cliente nÃ£o encontrado")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cliente)
}

func (h *ClienteHandler) Update(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		responderErro(w, http.StatusBadRequest, "ID obrigatÃ³rio")
		return
	}

	var req domain.Cliente
	if err := decodificarJSONLimitado(w, r, &req, 64*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Dados invÃ¡lidos")
		return
	}

	req.ID = id
	req.ContaID = usuario.ContaID

	if err := h.service.Update(r.Context(), &req, usuario.ID, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func (h *ClienteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		responderErro(w, http.StatusBadRequest, "ID obrigatÃ³rio")
		return
	}

	if err := h.service.Delete(r.Context(), id, usuario.ContaID, usuario.ID, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
