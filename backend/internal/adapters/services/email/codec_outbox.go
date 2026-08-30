package email

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type codecOutbox struct {
	segredos         ports.ProtetorSegredos
	maximoTentativas int
}

func NewCodecOutbox(segredos ports.ProtetorSegredos, maximoTentativas int) ports.PreparadorEmailOutbox {
	return &codecOutbox{segredos: segredos, maximoTentativas: maximoTentativas}
}

func (c *codecOutbox) PrepararEmail(
	contaID *string,
	chaveIdempotencia, destinatario, assunto, corpoHTML string,
) (*domain.EventoOutbox, error) {
	destinatario = strings.TrimSpace(destinatario)
	endereco, err := mail.ParseAddress(destinatario)
	if err != nil || !strings.EqualFold(endereco.Address, destinatario) {
		return nil, errors.New("destinatário de e-mail inválido")
	}
	assunto = strings.TrimSpace(assunto)
	if assunto == "" || len([]rune(assunto)) > 200 {
		return nil, errors.New("assunto de e-mail inválido")
	}
	chaveIdempotencia = strings.TrimSpace(chaveIdempotencia)
	if chaveIdempotencia == "" || len(chaveIdempotencia) > 200 {
		return nil, errors.New("chave de idempotência da notificação inválida")
	}
	if strings.TrimSpace(corpoHTML) == "" {
		return nil, errors.New("corpo do e-mail não informado")
	}
	payload, err := json.Marshal(domain.MensagemEmail{
		IDMensagem:   fmt.Sprintf("<%x@outbox.kaptei>", sha256.Sum256([]byte(chaveIdempotencia))),
		Destinatario: destinatario,
		Assunto:      assunto,
		CorpoHTML:    corpoHTML,
	})
	if err != nil {
		return nil, errors.New("não foi possível preparar o e-mail")
	}
	protegido, err := c.segredos.Proteger(string(payload))
	if err != nil {
		return nil, errors.New("não foi possível proteger a notificação")
	}
	agora := time.Now().UTC()
	return &domain.EventoOutbox{
		ContaID:           contaID,
		Tipo:              domain.TipoEventoEnviarEmail,
		PayloadProtegido:  protegido,
		ChaveIdempotencia: chaveIdempotencia,
		Status:            "PENDENTE",
		MaximoTentativas:  c.maximoTentativas,
		DisponivelEm:      agora,
		CriadoEm:          agora,
	}, nil
}

func (c *codecOutbox) DecodificarEmail(evento *domain.EventoOutbox) (*domain.MensagemEmail, error) {
	if evento == nil || evento.Tipo != domain.TipoEventoEnviarEmail {
		return nil, errors.New("evento de e-mail inválido")
	}
	payload, err := c.segredos.Revelar(evento.PayloadProtegido)
	if err != nil {
		return nil, errors.New("não foi possível revelar a notificação")
	}
	var mensagem domain.MensagemEmail
	if err := json.Unmarshal([]byte(payload), &mensagem); err != nil {
		return nil, errors.New("payload da notificação inválido")
	}
	return &mensagem, nil
}
