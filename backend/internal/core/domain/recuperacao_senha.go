package domain

import (
	"time"
)

// RecuperacaoSenhaToken representa o token gerado para recuperação de senha
type RecuperacaoSenhaToken struct {
	ID        string    `json:"id"`
	UsuarioID string    `json:"usuario_id"`
	Token     string    `json:"token"`
	ExpiraEm  time.Time `json:"expira_em"`
	Usado     bool      `json:"usado"`
	CriadoEm  time.Time `json:"criado_em"`
}

// TokenExpirado verifica se o token já passou da data de expiração
func (t *RecuperacaoSenhaToken) TokenExpirado() bool {
	return time.Now().After(t.ExpiraEm)
}
