package domain

import "time"

type SessaoUsuario struct {
	ID        string
	UsuarioID string
	ContaID   string
	ExpiraEm  time.Time
	CriadoEm  time.Time
}
