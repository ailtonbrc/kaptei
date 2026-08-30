package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type contextKey string

const UserContextKey contextKey = "user"
const CookieSessao = "kaptei_sessao"

func RequireAuth(authService ports.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := TokenDaRequisicao(r)
			if token == "" {
				responderErroMiddleware(w, http.StatusUnauthorized, "não autorizado: sessão ausente")
				return
			}

			usuario, err := authService.ValidateToken(r.Context(), token)
			if err != nil || usuario == nil {
				responderErroMiddleware(w, http.StatusUnauthorized, "não autorizado: sessão inválida ou expirada")
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, usuario)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TokenDaRequisicao(r *http.Request) string {
	if cookie, err := r.Cookie(CookieSessao); err == nil && strings.TrimSpace(cookie.Value) != "" {
		return strings.TrimSpace(cookie.Value)
	}
	autorizacao := r.Header.Get("Authorization")
	if strings.HasPrefix(autorizacao, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(autorizacao, "Bearer "))
	}
	return ""
}

func RequireRole(role domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			usuario, ok := r.Context().Value(UserContextKey).(*domain.Usuario)
			if !ok {
				responderErroMiddleware(w, http.StatusUnauthorized, "não autorizado: contexto de usuário inválido")
				return
			}
			if usuario.Papel == domain.RoleSuperAdmin || usuario.Papel == role {
				next.ServeHTTP(w, r)
				return
			}
			responderErroMiddleware(w, http.StatusForbidden, "proibido: nível de acesso insuficiente")
		})
	}
}

func RequireActivePlan(contaRepo ports.ContaRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			usuario, ok := r.Context().Value(UserContextKey).(*domain.Usuario)
			if !ok {
				responderErroMiddleware(w, http.StatusUnauthorized, "não autorizado: contexto de usuário inválido")
				return
			}
			if usuario.Papel == domain.RoleSuperAdmin {
				next.ServeHTTP(w, r)
				return
			}

			conta, err := contaRepo.GetByID(r.Context(), usuario.ContaID)
			if err != nil || conta == nil {
				responderErroMiddleware(w, http.StatusInternalServerError, "erro ao validar o status da conta")
				return
			}
			if conta.StatusPlano == "TRIAL" && conta.TrialVenceEm != nil {
				if time.Now().After(*conta.TrialVenceEm) {
					responderErroMiddleware(w, http.StatusPaymentRequired, "período gratuito encerrado; ative seu plano")
					return
				}
			} else if conta.StatusPlano != "ATIVO" {
				responderErroMiddleware(w, http.StatusPaymentRequired, "assinatura inativa; efetue o pagamento para liberar o acesso")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
