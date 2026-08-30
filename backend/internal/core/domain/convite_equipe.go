package domain

import "time"

type ConviteEquipe struct {
	ID           string     `json:"id"`
	ContaID      string     `json:"-"`
	Email        string     `json:"email"`
	Papel        Role       `json:"papel"`
	TokenHash    string     `json:"-"`
	ConvidadoPor string     `json:"-"`
	ExpiraEm     time.Time  `json:"expira_em"`
	UsadoEm      *time.Time `json:"usado_em,omitempty"`
	RevogadoEm   *time.Time `json:"revogado_em,omitempty"`
	CriadoEm     time.Time  `json:"criado_em"`
}
