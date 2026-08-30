package handlers

import (
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type UsuarioHandler struct {
	servico ports.UsuarioService
}

func NewUsuarioHandler(servico ports.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{servico: servico}
}

func (h *UsuarioHandler) UpdatePerfil(w http.ResponseWriter, r *http.Request) {
	// Extrair o usuÃ¡rio do Contexto (inserido pelo JWT Middleware)
	userCtx, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || userCtx == nil {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	var dto domain.AtualizacaoPerfil
	if err := decodificarJSONLimitado(w, r, &dto, 32*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Payload invÃ¡lido")
		return
	}

	usuario, err := h.servico.AtualizarPerfil(r.Context(), userCtx.ID, dto)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]interface{}{
		"mensagem": "Perfil atualizado com sucesso",
		"usuario":  usuario,
	})
}

// UpdateSenhaDTO representa o payload para alterar a senha
type UpdateSenhaDTO struct {
	SenhaAtual string `json:"senha_atual"`
	NovaSenha  string `json:"nova_senha"`
}

func (h *UsuarioHandler) UpdateSenha(w http.ResponseWriter, r *http.Request) {
	// Extrair o usuÃ¡rio do Contexto (inserido pelo JWT Middleware)
	userCtx, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || userCtx == nil {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	var dto UpdateSenhaDTO
	if err := decodificarJSONLimitado(w, r, &dto, 16*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Payload invÃ¡lido")
		return
	}

	if err := h.servico.AlterarSenha(r.Context(), userCtx.ID, dto.SenhaAtual, dto.NovaSenha); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "Senha alterada com sucesso; entre novamente."})
}
