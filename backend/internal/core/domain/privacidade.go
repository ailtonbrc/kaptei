package domain

import "time"

type TipoSolicitacaoTitular string

const (
	TipoConfirmacao                TipoSolicitacaoTitular = "CONFIRMACAO"
	TipoAcesso                     TipoSolicitacaoTitular = "ACESSO"
	TipoCorrecao                   TipoSolicitacaoTitular = "CORRECAO"
	TipoAnonimizacao               TipoSolicitacaoTitular = "ANONIMIZACAO"
	TipoBloqueio                   TipoSolicitacaoTitular = "BLOQUEIO"
	TipoExclusao                   TipoSolicitacaoTitular = "EXCLUSAO"
	TipoPortabilidade              TipoSolicitacaoTitular = "PORTABILIDADE"
	TipoRevogacao                  TipoSolicitacaoTitular = "REVOGACAO"
	TipoInformacaoCompartilhamento TipoSolicitacaoTitular = "INFORMACAO_COMPARTILHAMENTO"
)

type NovaSolicitacaoTitular struct {
	Tipo     TipoSolicitacaoTitular `json:"tipo"`
	Nome     string                 `json:"nome"`
	Email    string                 `json:"email"`
	Telefone string                 `json:"telefone"`
	Detalhes string                 `json:"detalhes"`
}

type SolicitacaoTitular struct {
	ID                     string                     `json:"id"`
	ContaID                string                     `json:"conta_id"`
	Protocolo              string                     `json:"protocolo"`
	Tipo                   TipoSolicitacaoTitular     `json:"tipo"`
	Nome                   string                     `json:"nome,omitempty"`
	Email                  string                     `json:"email,omitempty"`
	Telefone               string                     `json:"telefone,omitempty"`
	Detalhes               string                     `json:"detalhes,omitempty"`
	Status                 string                     `json:"status"`
	PrazoRespostaEm        time.Time                  `json:"prazo_resposta_em"`
	IdentidadeVerificadaEm *time.Time                 `json:"identidade_verificada_em,omitempty"`
	MetodoVerificacao      *string                    `json:"metodo_verificacao,omitempty"`
	Decisao                *string                    `json:"decisao,omitempty"`
	FundamentoLegal        *string                    `json:"fundamento_legal,omitempty"`
	ObservacaoDecisao      *string                    `json:"observacao_decisao,omitempty"`
	DecididaEm             *time.Time                 `json:"decidida_em,omitempty"`
	ExecutadaEm            *time.Time                 `json:"executada_em,omitempty"`
	CriadoEm               time.Time                  `json:"criado_em"`
	AtualizadoEm           time.Time                  `json:"atualizado_em"`
	NomeProtegido          string                     `json:"-"`
	ContatoProtegido       string                     `json:"-"`
	DetalhesProtegidos     *string                    `json:"-"`
	EmailHash              *string                    `json:"-"`
	TelefoneHash           *string                    `json:"-"`
	Eventos                []EventoSolicitacaoTitular `json:"eventos,omitempty"`
}

type DadosTitularPersistidos struct {
	Clientes     []map[string]any `json:"clientes"`
	Leads        []map[string]any `json:"leads"`
	Interacoes   []map[string]any `json:"interacoes"`
	Agendamentos []map[string]any `json:"agendamentos"`
	Conversas    []map[string]any `json:"conversas_whatsapp"`
	Mensagens    []map[string]any `json:"mensagens_whatsapp"`
}

type DecisaoSolicitacaoTitular struct {
	Aprovada        bool   `json:"aprovada"`
	FundamentoLegal string `json:"fundamento_legal"`
	Observacao      string `json:"observacao"`
}

type EventoSolicitacaoTitular struct {
	ID        string    `json:"id"`
	Tipo      string    `json:"tipo"`
	Descricao string    `json:"descricao"`
	UsuarioID *string   `json:"usuario_id,omitempty"`
	CriadoEm  time.Time `json:"criado_em"`
}

type ExportacaoTitular struct {
	Protocolo string         `json:"protocolo"`
	GeradaEm  time.Time      `json:"gerada_em"`
	Dados     map[string]any `json:"dados"`
}
