package domain

import "time"

type DominioSite struct {
	ID                  string     `json:"id"`
	ContaID             string     `json:"-"`
	Hostname            string     `json:"hostname"`
	Status              string     `json:"status"`
	RegistroTXTNome     string     `json:"registro_txt_nome"`
	RegistroTXTValor    string     `json:"registro_txt_valor"`
	TokenVerificacao    string     `json:"-"`
	VerificadoEm        *time.Time `json:"verificado_em,omitempty"`
	UltimaVerificacaoEm *time.Time `json:"ultima_verificacao_em,omitempty"`
	UltimoErro          *string    `json:"ultimo_erro,omitempty"`
	CriadoEm            time.Time  `json:"criado_em"`
	AtualizadoEm        time.Time  `json:"atualizado_em"`
}
