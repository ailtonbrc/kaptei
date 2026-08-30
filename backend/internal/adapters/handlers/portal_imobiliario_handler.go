package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"github.com/msdev/kaptei/internal/core/services"
)

type PortalImobiliarioHandler struct {
	servico ports.PortalImobiliarioService
}

func NewPortalImobiliarioHandler(servico ports.PortalImobiliarioService) *PortalImobiliarioHandler {
	return &PortalImobiliarioHandler{servico: servico}
}

func (h *PortalImobiliarioHandler) ObterConfiguracao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioPortal(w, r)
	if !ok {
		return
	}
	configuracao, err := h.servico.ObterConfiguracao(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusForbidden, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, configuracao)
}

func (h *PortalImobiliarioHandler) SalvarConfiguracao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioPortal(w, r)
	if !ok {
		return
	}
	var configuracao domain.ConfiguracaoPortal
	if decodificarJSONLimitado(w, r, &configuracao, 16*1024) != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "configuração inválida"})
		return
	}
	salva, err := h.servico.SalvarConfiguracao(r.Context(), usuario.ContaID, usuario.ID, usuario.Papel, configuracao)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, salva)
}

func (h *PortalImobiliarioHandler) RotacionarToken(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioPortal(w, r)
	if !ok {
		return
	}
	credencial, err := h.servico.RotacionarToken(r.Context(), usuario.ContaID, usuario.ID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	responderJSON(w, http.StatusCreated, credencial)
}

func (h *PortalImobiliarioHandler) ListarPublicacoes(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioPortal(w, r)
	if !ok {
		return
	}
	publicacoes, err := h.servico.ListarPublicacoes(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, publicacoes)
}

func (h *PortalImobiliarioHandler) SalvarPublicacao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioPortal(w, r)
	if !ok {
		return
	}
	var atualizacao domain.AtualizacaoPublicacaoPortal
	if decodificarJSONLimitado(w, r, &atualizacao, 4*1024) != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "publicação inválida"})
		return
	}
	if err := h.servico.SalvarPublicacao(r.Context(), usuario.ContaID, r.PathValue("id"), usuario.ID, usuario.Papel, atualizacao); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PortalImobiliarioHandler) Diagnosticar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioPortal(w, r)
	if !ok {
		return
	}
	diagnostico, err := h.servico.Diagnosticar(r.Context(), usuario.ContaID, usuario.Papel)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": err.Error()})
		return
	}
	responderJSON(w, http.StatusOK, diagnostico)
}

func (h *PortalImobiliarioHandler) FeedVRSync(w http.ResponseWriter, r *http.Request) {
	conteudo, err := h.servico.GerarFeedPublico(r.Context(), r.PathValue("token"))
	if err != nil {
		w.Header().Set("Cache-Control", "no-store")
		http.Error(w, "feed indisponível", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(conteudo)
}

func (h *PortalImobiliarioHandler) ReceberLead(w http.ResponseWriter, r *http.Request) {
	var lead domain.LeadGrupoOLX
	if decodificarJSONPortal(w, r, &lead, 64*1024) != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "payload de lead inválido"})
		return
	}
	err := h.servico.ReceberLead(r.Context(), r.PathValue("token"), r.Header.Get("Authorization"), lead)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
	case errors.Is(err, services.ErrPortalNaoAutorizado):
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
	case errors.Is(err, services.ErrLeadPortalInvalido):
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": err.Error()})
	case errors.Is(err, services.ErrPortalIndisponivel):
		responderJSON(w, http.StatusServiceUnavailable, map[string]string{"erro": "integração indisponível"})
	default:
		responderJSON(w, http.StatusInternalServerError, map[string]string{"erro": "não foi possível receber o lead"})
	}
}

func decodificarJSONPortal(w http.ResponseWriter, r *http.Request, destino any, limite int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, limite)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destino); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("corpo JSON deve conter um único objeto")
	}
	return nil
}

func usuarioPortal(w http.ResponseWriter, r *http.Request) (*domain.Usuario, bool) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario == nil {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return nil, false
	}
	return usuario, true
}
