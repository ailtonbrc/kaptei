package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type repositorioOutboxTeste struct {
	eventos         []*domain.EventoOutbox
	concluidos      []string
	falhos          []string
	falhaDefinitiva bool
}

func (r *repositorioOutboxTeste) Reservar(context.Context, string, int, time.Duration) ([]*domain.EventoOutbox, error) {
	return r.eventos, nil
}
func (r *repositorioOutboxTeste) Concluir(_ context.Context, eventoID, _ string) error {
	r.concluidos = append(r.concluidos, eventoID)
	return nil
}
func (r *repositorioOutboxTeste) Falhar(_ context.Context, eventoID, _, _ string, _ time.Time, definitivo bool) error {
	r.falhos = append(r.falhos, eventoID)
	r.falhaDefinitiva = definitivo
	return nil
}

type codecEmailTeste struct{ erro error }

func (c codecEmailTeste) PrepararEmail(*string, string, string, string, string) (*domain.EventoOutbox, error) {
	return nil, errors.New("não utilizado")
}

type codecObjetoTeste struct{ erro error }

func (c codecObjetoTeste) PrepararExclusaoObjeto(string, string, string, string) (*domain.EventoOutbox, error) {
	return nil, errors.New("não utilizado")
}
func (c codecObjetoTeste) DecodificarExclusaoObjeto(*domain.EventoOutbox) (*domain.SolicitacaoExclusaoObjeto, error) {
	if c.erro != nil {
		return nil, c.erro
	}
	return &domain.SolicitacaoExclusaoObjeto{Provedor: "teste", Chave: "conta/imagem.jpg"}, nil
}

type armazenamentoTeste struct{ exclusoes int }

func (a *armazenamentoTeste) Salvar(context.Context, string, []byte, string) (string, error) {
	return "", errors.New("não utilizado")
}
func (a *armazenamentoTeste) Excluir(context.Context, string) error {
	a.exclusoes++
	return nil
}
func (*armazenamentoTeste) Nome() string { return "teste" }
func (c codecEmailTeste) DecodificarEmail(*domain.EventoOutbox) (*domain.MensagemEmail, error) {
	if c.erro != nil {
		return nil, c.erro
	}
	return &domain.MensagemEmail{Destinatario: "pessoa@example.com", Assunto: "Assunto", CorpoHTML: "corpo"}, nil
}

type entregadorEmailTeste struct {
	erro   error
	envios int
}

func (e *entregadorEmailTeste) SendEmail(context.Context, string, string, string, string) error {
	e.envios++
	return e.erro
}

func novoProcessadorTeste(repo *repositorioOutboxTeste, codec codecEmailTeste, entregador *entregadorEmailTeste) *processadorOutbox {
	return NewProcessadorOutbox(repo, []ports.TratadorEventoOutbox{
		NewTratadorEmailOutbox(codec, entregador), NewTratadorObjetoOutbox(codecObjetoTeste{}, &armazenamentoTeste{}),
	}, ConfiguracaoProcessadorOutbox{
		TrabalhadorID: "teste", Intervalo: time.Second, TamanhoLote: 10,
		DuracaoBloqueio: time.Minute, BackoffInicial: time.Second, BackoffMaximo: time.Minute,
	})
}

func TestProcessadorOutboxExcluiObjeto(t *testing.T) {
	repo := &repositorioOutboxTeste{eventos: []*domain.EventoOutbox{{ID: "evento-objeto", Tipo: domain.TipoEventoExcluirObjeto, Tentativas: 1, MaximoTentativas: 3}}}
	objetos := &armazenamentoTeste{}
	processador := NewProcessadorOutbox(repo, []ports.TratadorEventoOutbox{
		NewTratadorEmailOutbox(codecEmailTeste{}, &entregadorEmailTeste{}), NewTratadorObjetoOutbox(codecObjetoTeste{}, objetos),
	}, ConfiguracaoProcessadorOutbox{
		TrabalhadorID: "teste", Intervalo: time.Second, TamanhoLote: 10,
		DuracaoBloqueio: time.Minute, BackoffInicial: time.Second, BackoffMaximo: time.Minute,
	})
	processador.processarLote(context.Background())
	if objetos.exclusoes != 1 || len(repo.concluidos) != 1 {
		t.Fatalf("objeto não foi excluído: exclusões=%d concluídos=%v", objetos.exclusoes, repo.concluidos)
	}
}

func TestProcessadorOutboxConcluiEmailEntregue(t *testing.T) {
	repo := &repositorioOutboxTeste{eventos: []*domain.EventoOutbox{{ID: "evento-1", Tipo: domain.TipoEventoEnviarEmail, Tentativas: 1, MaximoTentativas: 3}}}
	entregador := &entregadorEmailTeste{}
	novoProcessadorTeste(repo, codecEmailTeste{}, entregador).processarLote(context.Background())
	if entregador.envios != 1 || len(repo.concluidos) != 1 || len(repo.falhos) != 0 {
		t.Fatalf("resultado inesperado: envios=%d concluídos=%v falhos=%v", entregador.envios, repo.concluidos, repo.falhos)
	}
}

func TestProcessadorOutboxAgendaRetentativaDeFalhaTemporaria(t *testing.T) {
	repo := &repositorioOutboxTeste{eventos: []*domain.EventoOutbox{{ID: "evento-2", Tipo: domain.TipoEventoEnviarEmail, Tentativas: 2, MaximoTentativas: 3}}}
	entregador := &entregadorEmailTeste{erro: errors.New("SMTP indisponível")}
	novoProcessadorTeste(repo, codecEmailTeste{}, entregador).processarLote(context.Background())
	if len(repo.falhos) != 1 || repo.falhaDefinitiva {
		t.Fatalf("a falha temporária deveria ser reagendada: %+v", repo)
	}
}

func TestProcessadorOutboxEncerraEventoInvalido(t *testing.T) {
	repo := &repositorioOutboxTeste{eventos: []*domain.EventoOutbox{{ID: "evento-3", Tipo: "DESCONHECIDO", Tentativas: 1, MaximoTentativas: 8}}}
	novoProcessadorTeste(repo, codecEmailTeste{}, &entregadorEmailTeste{}).processarLote(context.Background())
	if len(repo.falhos) != 1 || !repo.falhaDefinitiva {
		t.Fatalf("evento inválido deveria falhar definitivamente: %+v", repo)
	}
}

func TestBackoffExponencialRespeitaLimite(t *testing.T) {
	processador := novoProcessadorTeste(&repositorioOutboxTeste{}, codecEmailTeste{}, &entregadorEmailTeste{})
	if atraso := processador.calcularBackoff(1); atraso != time.Second {
		t.Fatalf("backoff inicial = %s", atraso)
	}
	if atraso := processador.calcularBackoff(20); atraso != time.Minute {
		t.Fatalf("backoff máximo = %s", atraso)
	}
}
