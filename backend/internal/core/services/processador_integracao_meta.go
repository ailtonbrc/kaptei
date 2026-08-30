package services

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type ProcessadorIntegracaoMeta struct {
	repositorio ports.IntegracaoMetaRepository
	graph       ports.ClienteMetaGraph
	protetor    ports.ProtetorSegredos
	leads       ports.LeadService
	config      ConfiguracaoProcessadorOutbox
}

func NewProcessadorIntegracaoMeta(repositorio ports.IntegracaoMetaRepository, graph ports.ClienteMetaGraph, protetor ports.ProtetorSegredos, leads ports.LeadService, config ConfiguracaoProcessadorOutbox) *ProcessadorIntegracaoMeta {
	return &ProcessadorIntegracaoMeta{repositorio: repositorio, graph: graph, protetor: protetor, leads: leads, config: config}
}

func (p *ProcessadorIntegracaoMeta) Executar(ctx context.Context) {
	executarProcessadorFila(ctx, "meta_leads", p.config.Intervalo, p.processarLote)
}

func (p *ProcessadorIntegracaoMeta) processarLote(ctx context.Context) {
	processarLoteFila(ctx, operacoesFila[*domain.EventoIntegracao]{
		Nome: "meta_leads", Config: p.config, Reservar: p.repositorio.Reservar,
		Metadados: metadadosEventoIntegracao, Processar: p.processarEvento,
		Concluir: p.repositorio.Concluir, Falhar: p.repositorio.Falhar, Definitivo: erroNaoRetentavel,
	})
}

func (p *ProcessadorIntegracaoMeta) processarEvento(ctx context.Context, evento *domain.EventoIntegracao) error {
	configuracao, err := p.repositorio.ObterPorPagina(ctx, evento.PaginaID)
	if err != nil {
		return err
	}
	if configuracao == nil || configuracao.ContaID != evento.ContaID {
		return erroIntegracaoPermanente{"configuraÃ§Ã£o Meta ativa nÃ£o encontrada para o evento"}
	}
	token, err := p.protetor.Revelar(configuracao.TokenPaginaProtegido)
	if err != nil {
		return erroIntegracaoPermanente{"token da pÃ¡gina Meta nÃ£o pode ser revelado"}
	}
	leadMeta, err := p.graph.ObterLead(ctx, evento.IdentificadorExterno, token)
	if err != nil {
		return err
	}
	captura := mapearLeadMeta(leadMeta, evento)
	if err := p.leads.CaptarIntegracao(ctx, evento.ContaID, captura); err != nil {
		// Falhas de banco ou de distribuiÃ§Ã£o precisam ser retentadas. Dados
		// invÃ¡lidos esgotam o limite configurado e ficam visÃ­veis para operaÃ§Ã£o.
		return err
	}
	return nil
}

func mapearLeadMeta(lead *domain.LeadMeta, evento *domain.EventoIntegracao) domain.CapturaLeadIntegracao {
	nome := primeiroCampo(lead.Campos, "full_name")
	if nome == "" {
		nome = strings.TrimSpace(primeiroCampo(lead.Campos, "first_name") + " " + primeiroCampo(lead.Campos, "last_name"))
	}
	detalhes := []string{"Lead recebido via formulÃ¡rio Meta"}
	formularioID := lead.FormularioID
	if formularioID == nil {
		formularioID = evento.FormularioID
	}
	if formularioID != nil {
		detalhes = append(detalhes, "formulÃ¡rio "+*formularioID)
	}
	anuncioID := lead.AnuncioID
	if anuncioID == nil {
		anuncioID = evento.AnuncioID
	}
	if anuncioID != nil {
		detalhes = append(detalhes, "anÃºncio "+*anuncioID)
	}
	return domain.CapturaLeadIntegracao{
		Nome: nome, Email: primeiroCampo(lead.Campos, "email"),
		Telefone: primeiroCampo(lead.Campos, "phone_number"), Origem: domain.ProvedorMeta,
		Mensagem: strings.Join(detalhes, " | "), ChaveIdempotencia: uuidDeterministico("meta:" + evento.IdentificadorExterno),
	}
}

func primeiroCampo(campos map[string][]string, nome string) string {
	for _, valor := range campos[nome] {
		if texto := strings.TrimSpace(valor); texto != "" {
			return texto
		}
	}
	return ""
}

func uuidDeterministico(valor string) string {
	soma := sha256.Sum256([]byte(valor))
	soma[6] = (soma[6] & 0x0f) | 0x50
	soma[8] = (soma[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", soma[0:4], soma[4:6], soma[6:8], soma[8:10], soma[10:16])
}

type erroIntegracaoPermanente struct{ mensagem string }

func (e erroIntegracaoPermanente) Error() string       { return e.mensagem }
func (e erroIntegracaoPermanente) NaoRetentavel() bool { return true }

func erroNaoRetentavel(err error) bool {
	var erro ports.ErroNaoRetentavel
	return errors.As(err, &erro) && erro.NaoRetentavel()
}

func metadadosEventoIntegracao(evento *domain.EventoIntegracao) metadadosItemFila {
	return metadadosItemFila{ID: evento.ID, Tentativa: evento.Tentativas, MaximoTentativas: evento.MaximoTentativas}
}
