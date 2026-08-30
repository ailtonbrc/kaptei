package middlewares

import (
	"net/http"
	"strings"
)

// ValidarOrigemCookie protege mutações autenticadas por cookie contra CSRF.
// Clientes que usam Bearer sem cookie continuam aptos a integrar diretamente com a API.
func ValidarOrigemCookie(origensPermitidas []string) func(http.Handler) http.Handler {
	permitidas := make(map[string]struct{}, len(origensPermitidas))
	for _, origem := range origensPermitidas {
		permitidas[strings.TrimRight(strings.TrimSpace(origem), "/")] = struct{}{}
	}

	return func(proximo http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !metodoSeguro(r.Method) {
				if _, err := r.Cookie(CookieSessao); err == nil {
					origem := strings.TrimRight(strings.TrimSpace(r.Header.Get("Origin")), "/")
					if _, ok := permitidas[origem]; origem == "" || !ok {
						responderErroMiddleware(w, http.StatusForbidden, "origem da requisição não autorizada")
						return
					}
				}
			}
			proximo.ServeHTTP(w, r)
		})
	}
}

func metodoSeguro(metodo string) bool {
	return metodo == http.MethodGet || metodo == http.MethodHead || metodo == http.MethodOptions
}
