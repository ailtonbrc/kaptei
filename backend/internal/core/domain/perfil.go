package domain

type AtualizacaoPerfil struct {
	NomeCompleto     string  `json:"nome_completo"`
	Telefone         *string `json:"telefone"`
	NumeroWhatsapp   *string `json:"numero_whatsapp"`
	CPF              *string `json:"cpf"`
	RG               *string `json:"rg"`
	RGEstado         *string `json:"rg_estado"`
	RGOrgaoExpedidor *string `json:"rg_orgao_expedidor"`
	Nacionalidade    *string `json:"nacionalidade"`
	EstadoCivil      *string `json:"estado_civil"`
	Creci            *string `json:"creci"`
	CreciEstado      *string `json:"creci_estado"`
	CEP              *string `json:"cep"`
	Logradouro       *string `json:"logradouro"`
	Numero           *string `json:"numero"`
	Complemento      *string `json:"complemento"`
	Bairro           *string `json:"bairro"`
	Cidade           *string `json:"cidade"`
	Estado           *string `json:"estado"`
}
