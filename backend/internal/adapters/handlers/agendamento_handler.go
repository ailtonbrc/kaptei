package handlers

import (
	"net/http"
	"time"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type AgendamentoHandler struct {
	service ports.AgendamentoService
}

func NewAgendamentoHandler(service ports.AgendamentoService) *AgendamentoHandler {
	return &AgendamentoHandler{service: service}
}

func (h *AgendamentoHandler) Create(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}

	var agendamento domain.Agendamento
	if err := decodificarJSONLimitado(w, r, &agendamento, 32*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Dados invÃ¡lidos")
		return
	}

	agendamento.ContaID = usuario.ContaID
	if agendamento.UsuarioID == "" || !podeGerenciarEquipe(usuario) {
		agendamento.UsuarioID = usuario.ID
	}

	if err := h.service.Create(r.Context(), &agendamento, usuario.ID, usuario.Papel); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	responderJSON(w, http.StatusCreated, agendamento)
}

func (h *AgendamentoHandler) List(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}

	inicio, fim, err := periodoConsulta(r)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	var usuarioID *string
	if id := r.URL.Query().Get("usuario_id"); id != "" {
		usuarioID = &id
	}
	if !podeGerenciarEquipe(usuario) {
		id := usuario.ID
		usuarioID = &id
	}

	agendamentos, err := h.service.List(r.Context(), usuario.ContaID, usuarioID, inicio, fim)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	responderJSON(w, http.StatusOK, agendamentos)
}

func (h *AgendamentoHandler) Update(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}

	id := r.PathValue("id")
	existente, err := h.service.GetByID(r.Context(), id, usuario.ContaID)
	if err != nil || existente == nil {
		responderErro(w, http.StatusNotFound, "Agendamento nÃ£o encontrado")
		return
	}
	if !podeGerenciarEquipe(usuario) && existente.UsuarioID != usuario.ID {
		responderErro(w, http.StatusForbidden, "Sem permissÃ£o para alterar este agendamento")
		return
	}

	var alteracao domain.Agendamento
	if err := decodificarJSONLimitado(w, r, &alteracao, 32*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Dados invÃ¡lidos")
		return
	}
	alteracao.ID = existente.ID
	alteracao.ContaID = existente.ContaID
	alteracao.UsuarioID = existente.UsuarioID
	alteracao.ClienteID = existente.ClienteID
	alteracao.ImovelID = existente.ImovelID

	if err := h.service.Update(r.Context(), &alteracao); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	responderJSON(w, http.StatusOK, alteracao)
}

func (h *AgendamentoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	usuario, ok := usuarioAutenticado(w, r)
	if !ok {
		return
	}

	id := r.PathValue("id")
	existente, err := h.service.GetByID(r.Context(), id, usuario.ContaID)
	if err != nil || existente == nil {
		responderErro(w, http.StatusNotFound, "Agendamento nÃ£o encontrado")
		return
	}
	if !podeGerenciarEquipe(usuario) && existente.UsuarioID != usuario.ID {
		responderErro(w, http.StatusForbidden, "Sem permissÃ£o para excluir este agendamento")
		return
	}
	if err := h.service.Delete(r.Context(), id, usuario.ContaID); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func usuarioAutenticado(w http.ResponseWriter, r *http.Request) (*domain.Usuario, bool) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return nil, false
	}
	return usuario, true
}

func podeGerenciarEquipe(usuario *domain.Usuario) bool {
	return usuario.Papel == domain.RoleGestor || usuario.Papel == domain.RoleSuperAdmin
}

func periodoConsulta(r *http.Request) (time.Time, time.Time, error) {
	inicioTexto := r.URL.Query().Get("inicio")
	fimTexto := r.URL.Query().Get("fim")
	if inicioTexto == "" && fimTexto == "" {
		agora := time.Now()
		inicio := time.Date(agora.Year(), agora.Month(), 1, 0, 0, 0, 0, agora.Location())
		return inicio, inicio.AddDate(0, 1, 0), nil
	}

	inicio, err := time.Parse(time.RFC3339, inicioTexto)
	if err != nil {
		return time.Time{}, time.Time{}, erroPeriodoInvalido{}
	}
	fim, err := time.Parse(time.RFC3339, fimTexto)
	if err != nil || !fim.After(inicio) || fim.Sub(inicio) > 366*24*time.Hour {
		return time.Time{}, time.Time{}, erroPeriodoInvalido{}
	}
	return inicio, fim, nil
}

type erroPeriodoInvalido struct{}

func (erroPeriodoInvalido) Error() string {
	return "PerÃ­odo invÃ¡lido; informe inÃ­cio e fim em RFC3339"
}
