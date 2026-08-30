package domain

import "time"

type StatusAgendamento string
type TipoAgendamento string

const (
	StatusAgendamentoAgendado      StatusAgendamento = "AGENDADO"
	StatusAgendamentoConfirmado    StatusAgendamento = "CONFIRMADO"
	StatusAgendamentoConcluido     StatusAgendamento = "CONCLUIDO"
	StatusAgendamentoCancelado     StatusAgendamento = "CANCELADO"
	StatusAgendamentoNaoCompareceu StatusAgendamento = "NAO_COMPARECEU"

	TipoAgendamentoVisita        TipoAgendamento = "VISITA"
	TipoAgendamentoLigacao       TipoAgendamento = "LIGACAO"
	TipoAgendamentoReuniaoOnline TipoAgendamento = "REUNIAO_ONLINE"
	TipoAgendamentoTarefa        TipoAgendamento = "TAREFA"
)

type Agendamento struct {
	ID             string            `json:"id"`
	ContaID        string            `json:"conta_id"`
	UsuarioID      string            `json:"usuario_id"`
	ClienteID      *string           `json:"cliente_id"`
	ImovelID       *string           `json:"imovel_id"`
	Titulo         string            `json:"titulo"`
	Descricao      string            `json:"descricao"`
	DataHoraInicio time.Time         `json:"data_hora_inicio"`
	DataHoraFim    time.Time         `json:"data_hora_fim"`
	Status         StatusAgendamento `json:"status"`
	Tipo           TipoAgendamento   `json:"tipo"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	ClienteNome    *string           `json:"cliente_nome,omitempty"`
	ImovelTitulo   *string           `json:"imovel_titulo,omitempty"`
}
