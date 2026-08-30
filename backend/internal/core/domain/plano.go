package domain

import (
	"encoding/json"
	"time"
)

type Plano struct {
	ID               string          `json:"id"`
	Codigo           string          `json:"codigo"`
	Tipo             string          `json:"tipo"`
	Nome             string          `json:"nome"`
	Subtitle         *string         `json:"subtitle,omitempty"`
	Preco            float64         `json:"preco"`
	Cor              string          `json:"cor"`
	Recomendado      bool            `json:"recomendado"`
	Features         json.RawMessage `json:"features"`
	Missing          json.RawMessage `json:"missing"`
	Ativo            bool            `json:"ativo"`
	CriadoEm         time.Time       `json:"criado_em"`
	AtualizadoEm     time.Time       `json:"atualizado_em"`
	GatewayPriceID   *string         `json:"-"`
	LimiteCorretores *int            `json:"limite_corretores,omitempty"`
}
