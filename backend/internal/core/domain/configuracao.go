package domain

import (
	"encoding/json"
	"time"
)

// ConfiguracaoSistema representa uma configuração genérica no banco de dados
type ConfiguracaoSistema struct {
	Chave        string          `json:"chave"`
	Valor        json.RawMessage `json:"valor"`
	Descricao    *string         `json:"descricao"`
	AtualizadoEm time.Time       `json:"atualizado_em"`
}

// SMTPConfig define a estrutura esperada para as credenciais SMTP salvas no banco
type SMTPConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}
