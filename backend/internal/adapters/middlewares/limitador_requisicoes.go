package middlewares

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type janelaRequisicoes struct {
	inicio     time.Time
	quantidade int
}

func LimitarRequisicoes(limite int, janela time.Duration, confiarProxy bool) func(http.Handler) http.Handler {
	var mutex sync.Mutex
	clientes := make(map[string]janelaRequisicoes)

	return func(proximo http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			agora := time.Now()
			chave := identificarCliente(r, confiarProxy)
			mutex.Lock()
			if _, existe := clientes[chave]; !existe && len(clientes) >= 100_000 {
				mutex.Unlock()
				responderErroMiddleware(w, http.StatusServiceUnavailable, "limite global temporariamente atingido")
				return
			}
			registro := clientes[chave]
			if registro.inicio.IsZero() || agora.Sub(registro.inicio) >= janela {
				registro = janelaRequisicoes{inicio: agora}
			}
			registro.quantidade++
			clientes[chave] = registro
			permitido := registro.quantidade <= limite
			if len(clientes) > 10_000 {
				for cliente, valor := range clientes {
					if agora.Sub(valor.inicio) >= janela {
						delete(clientes, cliente)
					}
				}
			}
			mutex.Unlock()

			if !permitido {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("Retry-After", strconv.FormatInt(max(1, int64(janela/time.Second)), 10))
				responderErroMiddleware(w, http.StatusTooManyRequests, "muitas solicitações; tente novamente mais tarde")
				return
			}
			proximo.ServeHTTP(w, r)
		})
	}
}

func identificarCliente(r *http.Request, confiarProxy bool) string {
	if confiarProxy {
		if encaminhado := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); encaminhado != "" {
			return encaminhado
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
