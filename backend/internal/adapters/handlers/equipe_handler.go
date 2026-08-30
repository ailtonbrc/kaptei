package handlers

import (
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type EquipeHandler struct{ servico ports.EquipeService }

func NewEquipeHandler(servico ports.EquipeService) *EquipeHandler {
	return &EquipeHandler{servico: servico}
}

type membroEquipeResponse struct {
	ID        string      `json:"id"`
	Nome      string      `json:"nome"`
	Email     string      `json:"email"`
	Papel     domain.Role `json:"papel"`
	Status    string      `json:"status"`
	URLAvatar *string     `json:"url_avatar,omitempty"`
}

func (h *EquipeHandler) Listar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioEquipeAutenticado(r)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	membros, convites, err := h.servico.Listar(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	resumo := make([]membroEquipeResponse, 0, len(membros))
	for _, membro := range membros {
		resumo = append(resumo, membroEquipeResponse{ID: membro.ID, Nome: membro.NomeCompleto, Email: membro.Email, Papel: membro.Papel, Status: membro.Status, URLAvatar: membro.URLAvatar})
	}
	responderJSON(w, http.StatusOK, map[string]any{"membros": resumo, "convites": convites})
}

func (h *EquipeHandler) Convidar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioEquipeAutenticado(r)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	if err := decodificarJSONLimitado(w, r, &req, 8<<10); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados do convite inválidos"})
		return
	}
	if err := h.servico.Convidar(r.Context(), usuario.ContaID, usuario.ID, req.Email, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusCreated, map[string]string{"mensagem": "convite enviado com sucesso"})
}

func (h *EquipeHandler) CancelarConvite(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioEquipeAutenticado(r)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	if err := h.servico.CancelarConvite(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *EquipeHandler) AtualizarStatus(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioEquipeAutenticado(r)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := decodificarJSONLimitado(w, r, &req, 8<<10); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "status inválido"})
		return
	}
	if err := h.servico.AtualizarStatus(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, req.Status, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "status atualizado"})
}

func (h *EquipeHandler) AceitarConvite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
		Nome  string `json:"nome"`
		Senha string `json:"senha"`
	}
	if err := decodificarJSONLimitado(w, r, &req, 16<<10); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados inválidos"})
		return
	}
	if err := h.servico.AceitarConvite(r.Context(), req.Token, req.Nome, req.Senha); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusCreated, map[string]string{"mensagem": "convite aceito; faça login para continuar"})
}

func usuarioEquipeAutenticado(r *http.Request) (*domain.Usuario, bool) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	return usuario, ok && usuario != nil
}
