package domain

import "time"

const TipoEventoEnviarEmail = "EMAIL_ENVIAR"
const TipoEventoExcluirObjeto = "OBJETO_EXCLUIR"
const TipoEventoEnviarWhatsApp = "WHATSAPP_ENVIAR"

type EventoOutbox struct {
	ID                string
	ContaID           *string
	Tipo              string
	PayloadProtegido  string
	ChaveIdempotencia string
	Status            string
	Tentativas        int
	MaximoTentativas  int
	DisponivelEm      time.Time
	BloqueadoAte      *time.Time
	BloqueadoPor      *string
	CriadoEm          time.Time
}

type MensagemEmail struct {
	IDMensagem   string `json:"id_mensagem"`
	Destinatario string `json:"destinatario"`
	Assunto      string `json:"assunto"`
	CorpoHTML    string `json:"corpo_html"`
}

type SolicitacaoExclusaoObjeto struct {
	Provedor string `json:"provedor"`
	Chave    string `json:"chave"`
}
