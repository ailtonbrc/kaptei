package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/msdev/kaptei/internal/core/ports"
)

var formatoRequestID = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,80}$`)

type respostaMonitorada struct {
	http.ResponseWriter
	status, bytes int
}

func (w *respostaMonitorada) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *respostaMonitorada) Write(dados []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(dados)
	w.bytes += n
	return n, err
}

func Observabilidade(proximo http.Handler, metricas ports.MetricasAplicacao) http.Handler {
	if metricas == nil {
		metricas = ports.MetricasNulas{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if !formatoRequestID.MatchString(requestID) {
			requestID = novoRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		monitor := &respostaMonitorada{ResponseWriter: w}
		defer func() {
			if recuperado := recover(); recuperado != nil {
				slog.Error("panic na requisição", "request_id", requestID, "erro", recuperado, "stack", string(debug.Stack()))
				if monitor.status == 0 {
					responderErroMiddleware(monitor, http.StatusInternalServerError, "erro interno")
				}
			}
			status := monitor.status
			if status == 0 {
				status = http.StatusOK
			}
			slog.Info("requisição HTTP", "request_id", requestID, "metodo", r.Method, "rota", rotaNormalizada(r),
				"status", status, "bytes", monitor.bytes, "duracao_ms", time.Since(inicio).Milliseconds())
			metricas.RegistrarHTTP(r.Context(), r.Method, rotaNormalizada(r), status, time.Since(inicio))
		}()
		proximo.ServeHTTP(monitor, r)
	})
}

// rotaNormalizada evita persistir IDs, tokens e outros valores dinâmicos presentes na URL.
func rotaNormalizada(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	return r.URL.Path
}

func novoRequestID() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "indisponivel"
	}
	return hex.EncodeToString(bytes)
}
