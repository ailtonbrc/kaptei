package services

import (
	"context"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type ProcessadorIntegracaoWhatsApp struct {
	repositorio ports.IntegracaoWhatsAppRepository
	preparador  ports.PreparadorWhatsAppIntegracao
	leads       ports.LeadService
	config      ConfiguracaoProcessadorOutbox
}

func NewProcessadorIntegracaoWhatsApp(repositorio ports.IntegracaoWhatsAppRepository, preparador ports.PreparadorWhatsAppIntegracao, leads ports.LeadService, config ConfiguracaoProcessadorOutbox) *ProcessadorIntegracaoWhatsApp {
	return &ProcessadorIntegracaoWhatsApp{repositorio: repositorio, preparador: preparador, leads: leads, config: config}
}

func (p *ProcessadorIntegracaoWhatsApp) Executar(ctx context.Context) {
	executarProcessadorFila(ctx, "whatsapp_entrada", p.config.Intervalo, p.processarLote)
}

func (p *ProcessadorIntegracaoWhatsApp) processarLote(ctx context.Context) {
	processarLoteFila(ctx, operacoesFila[*domain.EventoIntegracao]{
		Nome: "whatsapp_entrada", Config: p.config, Reservar: p.repositorio.Reservar,
		Metadados: metadadosEventoIntegracao, Processar: p.processarEvento,
		Concluir: p.repositorio.Concluir, Falhar: p.repositorio.Falhar, Definitivo: erroNaoRetentavel,
	})
}

func (p *ProcessadorIntegracaoWhatsApp) processarEvento(ctx context.Context, evento *domain.EventoIntegracao) error {
	mensagem, err := p.preparador.DecodificarMensagem(evento)
	if err != nil {
		return erroIntegracaoPermanente{"payload WhatsApp protegido inválido"}
	}
	configuracao, err := p.repositorio.ObterPorNumeroTelefone(ctx, mensagem.NumeroTelefoneID)
	if err != nil {
		return err
	}
	if configuracao == nil || configuracao.ContaID != evento.ContaID {
		return erroIntegracaoPermanente{"configuração WhatsApp ativa não encontrada para o evento"}
	}
	chaveLead := uuidDeterministico("whatsapp:" + mensagem.NumeroContato)
	if err := p.leads.CaptarIntegracao(ctx, evento.ContaID, domain.CapturaLeadIntegracao{
		Nome: mensagem.NomeContato, Telefone: mensagem.NumeroContato, Origem: domain.ProvedorWhatsApp,
		Mensagem: "Contato iniciado pelo WhatsApp", ChaveIdempotencia: chaveLead,
	}); err != nil {
		return err
	}
	return p.repositorio.RegistrarMensagem(ctx, evento.ContaID, mensagem, evento.PayloadProtegido, chaveLead)
}
