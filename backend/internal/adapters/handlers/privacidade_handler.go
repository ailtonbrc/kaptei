package handlers

import (
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type PrivacidadeHandler struct{ servico ports.PrivacidadeService }

func NewPrivacidadeHandler(servico ports.PrivacidadeService) *PrivacidadeHandler {
	return &PrivacidadeHandler{servico: servico}
}

func (h *PrivacidadeHandler) CriarPublica(w http.ResponseWriter, r *http.Request) {
	var nova domain.NovaSolicitacaoTitular
	if err := decodificarJSONLimitado(w, r, &nova, 16*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "solicitação de privacidade inválida"})
		return
	}
	protocolo, err := h.servico.CriarPublica(r.Context(), r.PathValue("slug"), nova)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusAccepted, map[string]string{
		"protocolo": protocolo,
		"mensagem":  "Solicitação recebida. Guarde o protocolo para acompanhar o atendimento com o controlador.",
	})
}

func (h *PrivacidadeHandler) Listar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAtualPrivacidade(w, r)
	if !ok {
		return
	}
	resultado, err := h.servico.Listar(r.Context(), usuario.ContaID, usuario.Papel, filtroPaginacao(r))
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, resultado)
}

func (h *PrivacidadeHandler) Obter(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAtualPrivacidade(w, r)
	if !ok {
		return
	}
	resultado, err := h.servico.Obter(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	if resultado == nil {
		responderJSON(w, http.StatusNotFound, map[string]string{"erro": "solicitação não encontrada"})
		return
	}
	responderJSON(w, http.StatusOK, resultado)
}

func (h *PrivacidadeHandler) VerificarIdentidade(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAtualPrivacidade(w, r)
	if !ok {
		return
	}
	var requisicao struct {
		Metodo    string `json:"metodo"`
		Evidencia string `json:"evidencia"`
	}
	if err := decodificarJSONLimitado(w, r, &requisicao, 8*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "verificação inválida"})
		return
	}
	if err := h.servico.VerificarIdentidade(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, usuario.Papel, requisicao.Metodo, requisicao.Evidencia); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"status": "identidade_verificada"})
}

func (h *PrivacidadeHandler) Decidir(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAtualPrivacidade(w, r)
	if !ok {
		return
	}
	var decisao domain.DecisaoSolicitacaoTitular
	if err := decodificarJSONLimitado(w, r, &decisao, 8*1024); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "decisão inválida"})
		return
	}
	if err := h.servico.Decidir(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, usuario.Papel, decisao); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"status": "decisao_registrada"})
}

func (h *PrivacidadeHandler) Exportar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAtualPrivacidade(w, r)
	if !ok {
		return
	}
	exportacao, err := h.servico.Exportar(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="kaptei-dados-`+exportacao.Protocolo+`.json"`)
	w.Header().Set("Cache-Control", "no-store")
	responderJSON(w, http.StatusOK, exportacao)
}

func (h *PrivacidadeHandler) Executar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAtualPrivacidade(w, r)
	if !ok {
		return
	}
	if err := h.servico.Executar(r.Context(), r.PathValue("id"), usuario.ContaID, usuario.ID, usuario.Papel); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"status": "concluida"})
}

func usuarioAtualPrivacidade(w http.ResponseWriter, r *http.Request) (*domain.Usuario, bool) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario == nil {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return nil, false
	}
	return usuario, true
}
