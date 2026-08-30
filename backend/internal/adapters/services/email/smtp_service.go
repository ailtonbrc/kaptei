package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	mailer "github.com/wneessen/go-mail"
)

const prazoEnvioSMTP = 15 * time.Second

type smtpService struct {
	configRepo ports.ConfiguracaoRepository
	segredos   ports.ProtetorSegredos
}

func NewSMTPService(configRepo ports.ConfiguracaoRepository, segredos ports.ProtetorSegredos) ports.EmailService {
	return &smtpService{
		configRepo: configRepo,
		segredos:   segredos,
	}
}

func (s *smtpService) SendEmail(ctx context.Context, idMensagem, to, subject, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	to = strings.TrimSpace(to)
	endereco, err := mail.ParseAddress(to)
	if err != nil || !strings.EqualFold(endereco.Address, to) {
		return errors.New("destinatário de e-mail inválido")
	}
	if strings.TrimSpace(subject) == "" || len([]rune(subject)) > 200 {
		return errors.New("assunto de e-mail inválido")
	}
	// 1. Obter configurações de SMTP do banco de dados
	configDB, err := s.configRepo.Get(ctx, "SMTP_CONFIG")
	if err != nil {
		return err
	}
	if configDB == nil {
		return errors.New("configuração SMTP não encontrada no banco de dados")
	}

	var smtpConfig domain.SMTPConfig
	if err := json.Unmarshal(configDB.Valor, &smtpConfig); err != nil {
		return errors.New("erro ao fazer unmarshal das configurações SMTP")
	}
	smtpConfig.Password, err = s.segredos.Revelar(smtpConfig.Password)
	if err != nil {
		return err
	}

	// A biblioteca atual aplica deadline na conexão e em cada operação SMTP.
	mensagem := mailer.NewMsg()
	if err := mensagem.FromFormat(smtpConfig.FromName, smtpConfig.FromEmail); err != nil {
		return fmt.Errorf("configurar remetente SMTP: %w", err)
	}
	if err := mensagem.To(to); err != nil {
		return fmt.Errorf("configurar destinatário SMTP: %w", err)
	}
	mensagem.Subject(subject)
	if id := normalizarIDMensagem(idMensagem); id != "" {
		mensagem.SetMessageIDWithValue(id)
	}
	mensagem.SetBodyString(mailer.TypeTextHTML, body)

	opcoes := []mailer.Option{
		mailer.WithPort(smtpConfig.Port),
		mailer.WithTimeout(prazoEnvioSMTP),
		mailer.WithUsername(smtpConfig.User),
		mailer.WithPassword(smtpConfig.Password),
		mailer.WithSMTPAuth(mailer.SMTPAuthAutoDiscover),
	}
	if smtpConfig.Port == 465 {
		opcoes = append(opcoes, mailer.WithSSL())
	}
	cliente, err := mailer.NewClient(strings.TrimSpace(smtpConfig.Host), opcoes...)
	if err != nil {
		return fmt.Errorf("configurar cliente SMTP: %w", err)
	}

	ctxEnvio, cancelar := context.WithTimeout(ctx, prazoEnvioSMTP)
	defer cancelar()
	if err := cliente.DialAndSendWithContext(ctxEnvio, mensagem); err != nil {
		if ctxEnvio.Err() != nil {
			return fmt.Errorf("enviar e-mail SMTP: %w", ctxEnvio.Err())
		}
		return fmt.Errorf("enviar e-mail SMTP: %w", err)
	}
	return nil
}

func normalizarIDMensagem(valor string) string {
	valor = strings.TrimSpace(valor)
	if len(valor) >= 2 && strings.HasPrefix(valor, "<") && strings.HasSuffix(valor, ">") {
		return valor[1 : len(valor)-1]
	}
	return valor
}
