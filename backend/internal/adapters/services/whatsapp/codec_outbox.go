package whatsapp

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type CodecOutbox struct {
	protetor         ports.ProtetorSegredos
	maximoTentativas int
}

func NewCodecOutbox(protetor ports.ProtetorSegredos, maximoTentativas int) *CodecOutbox {
	return &CodecOutbox{protetor: protetor, maximoTentativas: maximoTentativas}
}

func (c *CodecOutbox) PrepararEnvio(solicitacao domain.SolicitacaoEnvioWhatsApp) (*domain.EventoOutbox, string, error) {
	if strings.TrimSpace(solicitacao.IDMensagem) == "" || strings.TrimSpace(solicitacao.ContaID) == "" || strings.TrimSpace(solicitacao.Destinatario) == "" {
		return nil, "", errors.New("mensagem WhatsApp de saída incompleta")
	}
	conteudo, err := json.Marshal(solicitacao)
	if err != nil {
		return nil, "", err
	}
	protegido, err := c.protetor.Proteger(string(conteudo))
	if err != nil {
		return nil, "", err
	}
	agora := time.Now().UTC()
	contaID := solicitacao.ContaID
	evento := &domain.EventoOutbox{ContaID: &contaID, Tipo: domain.TipoEventoEnviarWhatsApp,
		PayloadProtegido: protegido, ChaveIdempotencia: "whatsapp:enviar:" + solicitacao.IDMensagem,
		MaximoTentativas: c.maximoTentativas, DisponivelEm: agora, CriadoEm: agora}
	return evento, protegido, nil
}

func (c *CodecOutbox) DecodificarEnvio(evento *domain.EventoOutbox) (*domain.SolicitacaoEnvioWhatsApp, error) {
	if evento == nil || evento.Tipo != domain.TipoEventoEnviarWhatsApp || evento.PayloadProtegido == "" {
		return nil, errors.New("evento de envio WhatsApp inválido")
	}
	solicitacao, err := c.DecodificarConteudo(evento.PayloadProtegido)
	if err != nil {
		return nil, err
	}
	if evento.ContaID == nil || solicitacao.ContaID != *evento.ContaID || "whatsapp:enviar:"+solicitacao.IDMensagem != evento.ChaveIdempotencia {
		return nil, errors.New("payload de envio WhatsApp não corresponde ao evento")
	}
	return solicitacao, nil
}

func (c *CodecOutbox) DecodificarConteudo(conteudoProtegido string) (*domain.SolicitacaoEnvioWhatsApp, error) {
	if strings.TrimSpace(conteudoProtegido) == "" {
		return nil, errors.New("conteúdo de envio WhatsApp protegido é obrigatório")
	}
	conteudo, err := c.protetor.Revelar(conteudoProtegido)
	if err != nil {
		return nil, err
	}
	var solicitacao domain.SolicitacaoEnvioWhatsApp
	if err := json.Unmarshal([]byte(conteudo), &solicitacao); err != nil {
		return nil, err
	}
	return &solicitacao, nil
}
