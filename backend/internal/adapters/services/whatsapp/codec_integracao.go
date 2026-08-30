package whatsapp

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type CodecIntegracao struct {
	protetor         ports.ProtetorSegredos
	maximoTentativas int
}

func NewCodecIntegracao(protetor ports.ProtetorSegredos, maximoTentativas int) *CodecIntegracao {
	return &CodecIntegracao{protetor: protetor, maximoTentativas: maximoTentativas}
}

func (c *CodecIntegracao) PrepararMensagem(contaID string, mensagem domain.MensagemWhatsAppRecebida) (*domain.EventoIntegracao, error) {
	if strings.TrimSpace(contaID) == "" || strings.TrimSpace(mensagem.IdentificadorExterno) == "" {
		return nil, errors.New("conta e identificador da mensagem WhatsApp são obrigatórios")
	}
	conteudo, err := json.Marshal(mensagem)
	if err != nil {
		return nil, err
	}
	protegido, err := c.protetor.Proteger(string(conteudo))
	if err != nil {
		return nil, err
	}
	agora := time.Now().UTC()
	return &domain.EventoIntegracao{
		ContaID: contaID, Provedor: domain.ProvedorWhatsApp,
		Tipo:                 domain.TipoEventoWhatsAppMensagemEntrada,
		IdentificadorExterno: mensagem.IdentificadorExterno,
		PaginaID:             mensagem.NumeroTelefoneID, PayloadProtegido: protegido,
		MaximoTentativas: c.maximoTentativas, DisponivelEm: agora, CriadoEm: agora,
	}, nil
}

func (c *CodecIntegracao) DecodificarMensagem(evento *domain.EventoIntegracao) (*domain.MensagemWhatsAppRecebida, error) {
	if evento == nil || evento.Provedor != domain.ProvedorWhatsApp || evento.Tipo != domain.TipoEventoWhatsAppMensagemEntrada || evento.PayloadProtegido == "" {
		return nil, errors.New("evento WhatsApp inválido")
	}
	mensagem, err := c.DecodificarConteudo(evento.PayloadProtegido)
	if err != nil {
		return nil, err
	}
	if mensagem.IdentificadorExterno != evento.IdentificadorExterno || mensagem.NumeroTelefoneID != evento.PaginaID {
		return nil, errors.New("payload WhatsApp não corresponde ao evento")
	}
	return mensagem, nil
}

func (c *CodecIntegracao) DecodificarConteudo(conteudoProtegido string) (*domain.MensagemWhatsAppRecebida, error) {
	if strings.TrimSpace(conteudoProtegido) == "" {
		return nil, errors.New("conteúdo WhatsApp protegido é obrigatório")
	}
	conteudo, err := c.protetor.Revelar(conteudoProtegido)
	if err != nil {
		return nil, err
	}
	var mensagem domain.MensagemWhatsAppRecebida
	if err := json.Unmarshal([]byte(conteudo), &mensagem); err != nil {
		return nil, err
	}
	return &mensagem, nil
}
