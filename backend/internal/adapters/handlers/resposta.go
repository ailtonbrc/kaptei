package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func responderJSON(w http.ResponseWriter, status int, dados interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dados)
}

func responderErro(w http.ResponseWriter, status int, mensagem string) {
	if status >= http.StatusInternalServerError {
		slog.Error("falha interna no handler", "erro", mensagem)
		mensagem = "erro interno ao processar a solicitação"
	}
	responderJSON(w, status, map[string]string{"erro": mensagem})
}
