package services

import (
	"context"
	"fmt"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type tratadorEmailOutbox struct {
	codec ports.PreparadorEmailOutbox
	email ports.EmailService
}

func NewTratadorEmailOutbox(codec ports.PreparadorEmailOutbox, email ports.EmailService) ports.TratadorEventoOutbox {
	return &tratadorEmailOutbox{codec: codec, email: email}
}
func (*tratadorEmailOutbox) Tipo() string { return domain.TipoEventoEnviarEmail }
func (t *tratadorEmailOutbox) Processar(ctx context.Context, evento *domain.EventoOutbox) error {
	mensagem, err := t.codec.DecodificarEmail(evento)
	if err != nil {
		return fmt.Errorf("%w: %v", errEventoOutboxInvalido, err)
	}
	return t.email.SendEmail(ctx, mensagem.IDMensagem, mensagem.Destinatario, mensagem.Assunto, mensagem.CorpoHTML)
}

type tratadorObjetoOutbox struct {
	codec   ports.PreparadorObjetoOutbox
	objetos ports.ArmazenamentoObjetos
}

func NewTratadorObjetoOutbox(codec ports.PreparadorObjetoOutbox, objetos ports.ArmazenamentoObjetos) ports.TratadorEventoOutbox {
	return &tratadorObjetoOutbox{codec: codec, objetos: objetos}
}
func (*tratadorObjetoOutbox) Tipo() string { return domain.TipoEventoExcluirObjeto }
func (t *tratadorObjetoOutbox) Processar(ctx context.Context, evento *domain.EventoOutbox) error {
	solicitacao, err := t.codec.DecodificarExclusaoObjeto(evento)
	if err != nil {
		return fmt.Errorf("%w: %v", errEventoOutboxInvalido, err)
	}
	if solicitacao.Provedor != t.objetos.Nome() {
		return fmt.Errorf("provedor de armazenamento %q indisponível", solicitacao.Provedor)
	}
	return t.objetos.Excluir(ctx, solicitacao.Chave)
}

type tratadorWhatsAppOutbox struct {
	codec       ports.PreparadorWhatsAppOutbox
	cliente     ports.ClienteWhatsApp
	repositorio ports.IntegracaoWhatsAppRepository
	protetor    ports.ProtetorSegredos
}

func NewTratadorWhatsAppOutbox(codec ports.PreparadorWhatsAppOutbox, cliente ports.ClienteWhatsApp, repositorio ports.IntegracaoWhatsAppRepository, protetor ports.ProtetorSegredos) ports.TratadorEventoOutbox {
	return &tratadorWhatsAppOutbox{codec: codec, cliente: cliente, repositorio: repositorio, protetor: protetor}
}
func (*tratadorWhatsAppOutbox) Tipo() string { return domain.TipoEventoEnviarWhatsApp }
func (t *tratadorWhatsAppOutbox) Processar(ctx context.Context, evento *domain.EventoOutbox) error {
	solicitacao, err := t.codec.DecodificarEnvio(evento)
	if err != nil {
		return fmt.Errorf("%w: %v", errEventoOutboxInvalido, err)
	}
	if existente, err := t.repositorio.ObterIdentificadorMensagemSaida(ctx, solicitacao.IDMensagem); err != nil {
		return err
	} else if existente != nil {
		return nil
	}
	if t.cliente == nil {
		return fmt.Errorf("cliente WhatsApp Graph indisponível")
	}
	configuracao, err := t.repositorio.ObterPorConta(ctx, solicitacao.ContaID)
	if err != nil {
		return err
	}
	if configuracao == nil || !configuracao.Ativa || configuracao.NumeroTelefoneID != solicitacao.NumeroTelefoneID {
		return fmt.Errorf("%w: configuração WhatsApp ativa não encontrada", errEventoOutboxInvalido)
	}
	token, err := t.protetor.Revelar(configuracao.TokenAcessoProtegido)
	if err != nil {
		return fmt.Errorf("%w: token WhatsApp não pode ser revelado", errEventoOutboxInvalido)
	}
	identificador, err := t.cliente.Enviar(ctx, configuracao.NumeroTelefoneID, token, solicitacao)
	if err != nil {
		return err
	}
	return t.repositorio.MarcarMensagemEnviada(ctx, solicitacao.IDMensagem, identificador)
}
