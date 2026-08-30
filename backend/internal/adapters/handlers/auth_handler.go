package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

const duracaoSessao = 24 * time.Hour

type AuthHandler struct {
	authService  ports.AuthService
	contaRepo    ports.ContaRepository
	cookieSeguro bool
}

func NewAuthHandler(authService ports.AuthService, contaRepo ports.ContaRepository, cookieSeguro bool) *AuthHandler {
	return &AuthHandler{authService: authService, contaRepo: contaRepo, cookieSeguro: cookieSeguro}
}

type loginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type googleLoginRequest struct {
	Token     string `json:"token"`
	TipoConta string `json:"tipo_conta"`
	Plano     string `json:"plano"`
}

type registerRequest struct {
	Nome      string `json:"nome"`
	Email     string `json:"email"`
	Senha     string `json:"senha"`
	TipoConta string `json:"tipo_conta"`
	Plano     string `json:"plano"`
}

type sessaoResponse struct {
	ID           string      `json:"id"`
	Nome         string      `json:"nome"`
	Email        string      `json:"email"`
	Papel        domain.Role `json:"papel"`
	ContaID      string      `json:"conta_id"`
	Avatar       *string     `json:"avatar"`
	StatusPlano  string      `json:"status_plano"`
	Plano        string      `json:"plano"`
	TrialVenceEm *time.Time  `json:"trial_vence_em"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodificarJSONLimitado(w, r, &req, 32<<10); err != nil {
		return
	}
	token, err := h.authService.Login(r.Context(), strings.TrimSpace(req.Email), req.Senha)
	if err != nil {
		responderErro(w, http.StatusUnauthorized, "E-mail ou senha invÃ¡lidos.")
		return
	}
	h.responderAutenticacao(w, r, token)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req googleLoginRequest
	if err := decodificarJSONLimitado(w, r, &req, 64<<10); err != nil {
		return
	}
	token, err := h.authService.GoogleLogin(r.Context(), req.Token, req.TipoConta, req.Plano)
	if err != nil {
		responderErro(w, http.StatusUnauthorized, "NÃ£o foi possÃ­vel autenticar com o Google.")
		return
	}
	h.responderAutenticacao(w, r, token)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodificarJSONLimitado(w, r, &req, 32<<10); err != nil {
		return
	}
	if strings.TrimSpace(req.Nome) == "" || strings.TrimSpace(req.Email) == "" || len(req.Senha) < 12 || len([]byte(req.Senha)) > 72 || req.TipoConta == "" || req.Plano == "" {
		responderErro(w, http.StatusBadRequest, "Preencha todos os campos, selecione um plano e use uma senha entre 12 e 72 bytes.")
		return
	}
	token, err := h.authService.Register(r.Context(), req.Nome, req.Email, req.Senha, req.TipoConta, req.Plano)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	h.responderAutenticacao(w, r, token)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if token := middlewares.TokenDaRequisicao(r); token != "" {
		_ = h.authService.Logout(r.Context(), token)
	}
	h.definirCookie(w, "", -1)
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "SessÃ£o encerrada com sucesso."})
}

func (h *AuthHandler) responderAutenticacao(w http.ResponseWriter, r *http.Request, token string) {
	usuario, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil || usuario == nil {
		responderErro(w, http.StatusInternalServerError, "NÃ£o foi possÃ­vel criar a sessÃ£o.")
		return
	}
	h.definirCookie(w, token, int(duracaoSessao.Seconds()))
	h.responderSessao(w, r, usuario)
}

// Sessao revalida a identidade no backend e devolve apenas os dados necessÃ¡rios Ã  interface.
func (h *AuthHandler) Sessao(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok || usuario == nil {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado.")
		return
	}
	h.responderSessao(w, r, usuario)
}

func (h *AuthHandler) responderSessao(w http.ResponseWriter, r *http.Request, usuario *domain.Usuario) {
	conta, err := h.contaRepo.GetByID(r.Context(), usuario.ContaID)
	if err != nil || conta == nil {
		responderErro(w, http.StatusInternalServerError, "NÃ£o foi possÃ­vel carregar a conta da sessÃ£o.")
		return
	}
	responderJSON(w, http.StatusOK, sessaoResponse{
		ID: usuario.ID, Nome: usuario.NomeCompleto, Email: usuario.Email, Papel: usuario.Papel,
		ContaID: usuario.ContaID, Avatar: usuario.URLAvatar, StatusPlano: conta.StatusPlano,
		Plano: conta.Plano, TrialVenceEm: conta.TrialVenceEm,
	})
}

func (h *AuthHandler) definirCookie(w http.ResponseWriter, token string, maxAge int) {
	expiraEm := time.Now().Add(duracaoSessao)
	if maxAge < 0 {
		expiraEm = time.Unix(1, 0)
	}
	http.SetCookie(w, &http.Cookie{
		Name: middlewares.CookieSessao, Value: token, Path: "/", MaxAge: maxAge,
		Expires: expiraEm, HttpOnly: true, Secure: h.cookieSeguro, SameSite: http.SameSiteLaxMode,
	})
}

type esqueciSenhaRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) EsqueciSenha(w http.ResponseWriter, r *http.Request) {
	var req esqueciSenhaRequest
	if err := decodificarJSONLimitado(w, r, &req, 16<<10); err != nil {
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		responderErro(w, http.StatusBadRequest, "O e-mail Ã© obrigatÃ³rio.")
		return
	}
	if err := h.authService.SolicitarRecuperacaoSenha(r.Context(), req.Email); err != nil {
		slog.ErrorContext(r.Context(), "falha ao solicitar recuperaÃ§Ã£o de senha", "erro", err)
	}
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "Se o e-mail estiver cadastrado, vocÃª receberÃ¡ um link para redefinir a senha."})
}

type redefinirSenhaRequest struct {
	Token string `json:"token"`
	Senha string `json:"senha"`
}

func (h *AuthHandler) RedefinirSenha(w http.ResponseWriter, r *http.Request) {
	var req redefinirSenhaRequest
	if err := decodificarJSONLimitado(w, r, &req, 32<<10); err != nil {
		return
	}
	if strings.TrimSpace(req.Token) == "" || len(req.Senha) < 12 || len([]byte(req.Senha)) > 72 {
		responderErro(w, http.StatusBadRequest, "Token invÃ¡lido ou senha fora do intervalo permitido.")
		return
	}
	if err := h.authService.RedefinirSenha(r.Context(), req.Token, req.Senha); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "Senha redefinida com sucesso. VocÃª jÃ¡ pode fazer login."})
}
