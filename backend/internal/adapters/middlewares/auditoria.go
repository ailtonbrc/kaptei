package middlewares

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/msdev/kaptei/internal/core/domain"
)

type respostaAuditoria struct {
	http.ResponseWriter
	status int
}

func (w *respostaAuditoria) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *respostaAuditoria) Write(d []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(d)
}

func AuditarMutacoes(db *sql.DB) func(http.Handler) http.Handler {
	return func(proximo http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				proximo.ServeHTTP(w, r)
				return
			}
			monitor := &respostaAuditoria{ResponseWriter: w}
			proximo.ServeHTTP(monitor, r)
			usuario, ok := r.Context().Value(UserContextKey).(*domain.Usuario)
			if !ok || usuario == nil {
				return
			}
			status := monitor.status
			if status == 0 {
				status = http.StatusOK
			}
			rota := rotaNormalizada(r)
			_, err := db.ExecContext(r.Context(), `INSERT INTO auditoria_eventos (conta_id,usuario_id,request_id,metodo,rota,status_http) VALUES ($1,$2,$3,$4,$5,$6)`, usuario.ContaID, usuario.ID, w.Header().Get("X-Request-ID"), r.Method, rota, status)
			if err != nil {
				slog.Error("falha ao gravar auditoria", "erro", err, "rota", rota)
			}
		})
	}
}
