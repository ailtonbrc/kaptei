package domain

import "time"

type Interacao struct {
	ID         string    `json:"id"`
	ContaID    string    `json:"conta_id"`
	ClienteID  string    `json:"cliente_id"`
	CorretorID *string   `json:"corretor_id"`
	Tipo       string    `json:"tipo"` // LIGACAO, MENSAGEM, VISITA, PROPOSTA, ANOTACAO
	Descricao  string    `json:"descricao"`
	DataHora   time.Time `json:"data_hora"`
	CriadoEm   time.Time `json:"criado_em"`
}
