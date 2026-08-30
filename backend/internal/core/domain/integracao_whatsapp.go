package domain

import "time"

const (
	ProvedorWhatsApp                  = "WHATSAPP"
	TipoEventoWhatsAppMensagemEntrada = "MENSAGEM_RECEBIDA"
)

type ConfiguracaoWhatsApp struct {
	ID                     string    `json:"id,omitempty"`
	ContaID                string    `json:"-"`
	WABAID                 string    `json:"waba_id"`
	NumeroTelefoneID       string    `json:"numero_telefone_id"`
	NumeroExibicao         *string   `json:"numero_exibicao,omitempty"`
	TokenAcessoProtegido   string    `json:"-"`
	TokenAcessoConfigurado bool      `json:"token_acesso_configurado"`
	DisponivelNoServidor   bool      `json:"disponivel_no_servidor"`
	Ativa                  bool      `json:"ativa"`
	CriadoEm               time.Time `json:"criado_em,omitempty"`
	AtualizadoEm           time.Time `json:"atualizado_em,omitempty"`
}

type AtualizacaoWhatsApp struct {
	WABAID           string  `json:"waba_id"`
	NumeroTelefoneID string  `json:"numero_telefone_id"`
	NumeroExibicao   *string `json:"numero_exibicao,omitempty"`
	TokenAcesso      string  `json:"token_acesso,omitempty"`
	Ativa            bool    `json:"ativa"`
}

type MensagemWhatsAppRecebida struct {
	IdentificadorExterno string    `json:"identificador_externo"`
	NumeroTelefoneID     string    `json:"numero_telefone_id"`
	NumeroContato        string    `json:"numero_contato"`
	NomeContato          string    `json:"nome_contato,omitempty"`
	Tipo                 string    `json:"tipo"`
	Texto                string    `json:"texto"`
	OcorridaEm           time.Time `json:"ocorrida_em"`
}

type ConversaWhatsApp struct {
	ID                     string     `json:"id"`
	ContaID                string     `json:"-"`
	LeadID                 *string    `json:"lead_id,omitempty"`
	NumeroContato          string     `json:"numero_contato"`
	NomeContato            *string    `json:"nome_contato,omitempty"`
	ConsentimentoMarketing bool       `json:"consentimento_marketing"`
	JanelaAtendimentoAte   *time.Time `json:"janela_atendimento_ate,omitempty"`
	UltimaMensagemEm       time.Time  `json:"ultima_mensagem_em"`
}

type MensagemWhatsApp struct {
	ID                   string     `json:"id"`
	ConversaID           string     `json:"conversa_id"`
	IdentificadorExterno *string    `json:"identificador_externo,omitempty"`
	Direcao              string     `json:"direcao"`
	Tipo                 string     `json:"tipo"`
	ConteudoProtegido    string     `json:"-"`
	Conteudo             string     `json:"conteudo"`
	Status               string     `json:"status"`
	OcorridaEm           time.Time  `json:"ocorrida_em"`
	EnviadaEm            *time.Time `json:"enviada_em,omitempty"`
	EntregueEm           *time.Time `json:"entregue_em,omitempty"`
	LidaEm               *time.Time `json:"lida_em,omitempty"`
	FalhouEm             *time.Time `json:"falhou_em,omitempty"`
	ErroCodigo           *string    `json:"erro_codigo,omitempty"`
	ErroDetalhe          *string    `json:"erro_detalhe,omitempty"`
}

type SolicitacaoEnvioWhatsApp struct {
	IDMensagem       string   `json:"id_mensagem"`
	ContaID          string   `json:"conta_id"`
	ConversaID       string   `json:"conversa_id"`
	NumeroTelefoneID string   `json:"numero_telefone_id"`
	Destinatario     string   `json:"destinatario"`
	Tipo             string   `json:"tipo"`
	Texto            string   `json:"texto,omitempty"`
	TemplateNome     string   `json:"template_nome,omitempty"`
	TemplateIdioma   string   `json:"template_idioma,omitempty"`
	Parametros       []string `json:"parametros,omitempty"`
}

type EnvioTemplateWhatsApp struct {
	Nome       string   `json:"nome"`
	Idioma     string   `json:"idioma"`
	Parametros []string `json:"parametros,omitempty"`
}
