package domain

import (
	"encoding/json"
	"time"
)

type Role string

const (
	RoleSuperAdmin     Role = "SUPER_ADMIN"
	RoleGestor         Role = "GESTOR"
	RoleCorretorEquipe Role = "CORRETOR_EQUIPE"
	RoleCorretorSolo   Role = "CORRETOR_SOLO"
)

type ContaSaaS struct {
	ID           string          `json:"id"`
	TipoConta    string          `json:"tipo_conta"` // 'CORRETOR_SOLO' ou 'IMOBILIARIA'
	NomeConta    *string         `json:"nome_conta"`
	StatusPlano  string          `json:"status_plano"`
	Plano        string          `json:"plano"`
	TrialVenceEm *time.Time      `json:"trial_vence_em"`
	FeatureFlags json.RawMessage `json:"feature_flags"`

	LeadEstrategia        string     `json:"lead_estrategia"`
	LeadTokenIntegracao   *string    `json:"lead_token_integracao"`
	LeadTokenHash         *string    `json:"-"`
	LeadTokenPrefixo      *string    `json:"lead_token_prefixo,omitempty"`
	BillingCustomerID     *string    `json:"-"`
	BillingSubscriptionID *string    `json:"-"`
	BillingStatus         *string    `json:"billing_status,omitempty"`
	BillingPeriodoFim     *time.Time `json:"billing_periodo_fim,omitempty"`

	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type Usuario struct {
	ID           string  `json:"id"`
	ContaID      string  `json:"conta_id"`
	NomeCompleto string  `json:"nome_completo"`
	Email        string  `json:"email"`
	SenhaHash    *string `json:"-"` // Oculto na serialização JSON
	GoogleID     *string `json:"-"`
	Papel        Role    `json:"papel"`
	Status       string  `json:"status"`
	URLAvatar    *string `json:"url_avatar"`
	VersaoSessao int     `json:"-"`

	// Dados Pessoais / Profissionais
	CPF              *string `json:"cpf"`
	RG               *string `json:"rg"`
	RGEstado         *string `json:"rg_estado"`
	RGOrgaoExpedidor *string `json:"rg_orgao_expedidor"`
	Nacionalidade    *string `json:"nacionalidade"`
	EstadoCivil      *string `json:"estado_civil"`
	Creci            *string `json:"creci"`
	CreciEstado      *string `json:"creci_estado"`

	// Endereço e Contato
	CEP            *string `json:"cep"`
	Logradouro     *string `json:"logradouro"`
	Numero         *string `json:"numero"`
	Complemento    *string `json:"complemento"`
	Bairro         *string `json:"bairro"`
	Cidade         *string `json:"cidade"`
	Estado         *string `json:"estado"`
	Telefone       *string `json:"telefone"`
	NumeroWhatsapp *string `json:"numero_whatsapp"`

	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}
