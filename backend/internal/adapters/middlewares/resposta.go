package middlewares

import (
	"encoding/json"
	"net/http"
)

func responderErroMiddleware(w http.ResponseWriter, status int, mensagem string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"erro": mensagem})
}
