package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

var (
	ErrMetaIndisponivel       = errors.New("integraÃ§Ã£o Meta nÃ£o configurada no servidor")
	ErrMetaAssinaturaInvalida = errors.New("assinatura do webhook Meta invÃ¡lida")
	formatoIDMeta             = regexp.MustCompile(`^[0-9]{1,64}$`)
)

type integracaoMetaService struct {
	repositorio       ports.IntegracaoMetaRepository
	protetor          ports.ProtetorSegredos
	segredoAplicativo string
	tokenVerificacao  string
	maximoTentativas  int
}

func NewIntegracaoMetaService(
	repositorio ports.IntegracaoMetaRepository,
	protetor ports.ProtetorSegredos,
	segredoAplicativo, tokenVerificacao string,
	maximoTentativas int,
) ports.IntegracaoMetaService {
	return &integracaoMetaService{
		repositorio: repositorio, protetor: protetor,
		segredoAplicativo: strings.TrimSpace(segredoAplicativo),
		tokenVerificacao:  strings.TrimSpace(tokenVerificacao),
		maximoTentativas:  maximoTentativas,
	}
}

func (s *integracaoMetaService) ObterConfiguracao(ctx context.Context, contaID string, papel domain.Role) (*domain.ConfiguracaoMetaLeads, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissÃ£o para administrar a integraÃ§Ã£o Meta")
	}
	configuracao, err := s.repositorio.ObterPorConta(ctx, contaID)
	if err != nil {
		return nil, err
	}
	if configuracao == nil {
		return &domain.ConfiguracaoMetaLeads{ContaID: contaID, DisponivelNoServidor: s.disponivel()}, nil
	}
	configuracao.TokenPaginaProtegido = ""
	configuracao.DisponivelNoServidor = s.disponivel()
	return configuracao, nil
}

func (s *integracaoMetaService) SalvarConfiguracao(ctx context.Context, contaID string, papel domain.Role, atualizacao domain.AtualizacaoMetaLeads) (*domain.ConfiguracaoMetaLeads, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissÃ£o para administrar a integraÃ§Ã£o Meta")
	}
	if atualizacao.Ativa && !s.disponivel() {
		return nil, ErrMetaIndisponivel
	}
	paginaID := strings.TrimSpace(atualizacao.PaginaID)
	if !formatoIDMeta.MatchString(paginaID) {
		return nil, errors.New("identificador da pÃ¡gina Meta invÃ¡lido")
	}
	existente, err := s.repositorio.ObterPorConta(ctx, contaID)
	if err != nil {
		return nil, err
	}
	tokenProtegido := ""
	if existente != nil {
		tokenProtegido = existente.TokenPaginaProtegido
	}
	token := strings.TrimSpace(atualizacao.TokenPagina)
	if token != "" {
		if len(token) < 20 || len(token) > 4096 || strings.ContainsAny(token, "\r\n\t ") {
			return nil, errors.New("token de acesso da pÃ¡gina Meta invÃ¡lido")
		}
		tokenProtegido, err = s.protetor.Proteger(token)
		if err != nil {
			return nil, fmt.Errorf("proteger token da pÃ¡gina Meta: %w", err)
		}
	}
	if tokenProtegido == "" {
		return nil, errors.New("informe o token de acesso da pÃ¡gina Meta")
	}
	configuracao := &domain.ConfiguracaoMetaLeads{
		ContaID: contaID, PaginaID: paginaID, TokenPaginaProtegido: tokenProtegido,
		TokenPaginaConfigurado: true, Ativa: atualizacao.Ativa,
	}
	if err := s.repositorio.Salvar(ctx, configuracao); err != nil {
		return nil, err
	}
	configuracao.TokenPaginaProtegido = ""
	configuracao.DisponivelNoServidor = s.disponivel()
	return configuracao, nil
}

func (s *integracaoMetaService) disponivel() bool {
	return s.segredoAplicativo != "" && s.tokenVerificacao != ""
}

func (s *integracaoMetaService) VerificarWebhook(modo, token, desafio string) (string, error) {
	if s.tokenVerificacao == "" {
		return "", ErrMetaIndisponivel
	}
	if strings.TrimSpace(modo) != "subscribe" || desafio == "" || !compararSegredo(token, s.tokenVerificacao) {
		return "", errors.New("verificaÃ§Ã£o do webhook Meta recusada")
	}
	return desafio, nil
}

func (s *integracaoMetaService) ReceberWebhook(ctx context.Context, assinatura string, corpo []byte) error {
	if s.segredoAplicativo == "" {
		return ErrMetaIndisponivel
	}
	if !validarAssinaturaMeta(corpo, assinatura, s.segredoAplicativo) {
		return ErrMetaAssinaturaInvalida
	}
	var notificacao struct {
		Objeto   string `json:"object"`
		Entradas []struct {
			PaginaID   string `json:"id"`
			Alteracoes []struct {
				Campo string `json:"field"`
				Valor struct {
					LeadID       string `json:"leadgen_id"`
					PaginaID     string `json:"page_id"`
					FormularioID string `json:"form_id"`
					AnuncioID    string `json:"ad_id"`
					CriadoEm     int64  `json:"created_time"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(corpo, &notificacao); err != nil || notificacao.Objeto != "page" {
		return errors.New("payload do webhook Meta invÃ¡lido")
	}
	agora := time.Now().UTC()
	eventos := make([]*domain.EventoIntegracao, 0)
	for _, entrada := range notificacao.Entradas {
		for _, alteracao := range entrada.Alteracoes {
			if alteracao.Campo != "leadgen" {
				continue
			}
			paginaID := strings.TrimSpace(alteracao.Valor.PaginaID)
			if paginaID == "" {
				paginaID = strings.TrimSpace(entrada.PaginaID)
			}
			leadID := strings.TrimSpace(alteracao.Valor.LeadID)
			if !formatoIDMeta.MatchString(paginaID) || !formatoIDMeta.MatchString(leadID) {
				return errors.New("payload Meta contÃ©m identificadores invÃ¡lidos")
			}
			configuracao, err := s.repositorio.ObterPorPagina(ctx, paginaID)
			if err != nil {
				return err
			}
			if configuracao == nil {
				continue
			}
			evento := &domain.EventoIntegracao{
				ContaID: configuracao.ContaID, Provedor: domain.ProvedorMeta,
				Tipo: domain.TipoEventoMetaLeadGerado, IdentificadorExterno: leadID,
				PaginaID: paginaID, MaximoTentativas: s.maximoTentativas,
				DisponivelEm: agora, CriadoEm: agora,
			}
			if formatoIDMeta.MatchString(alteracao.Valor.FormularioID) {
				evento.FormularioID = &alteracao.Valor.FormularioID
			}
			if formatoIDMeta.MatchString(alteracao.Valor.AnuncioID) {
				evento.AnuncioID = &alteracao.Valor.AnuncioID
			}
			if alteracao.Valor.CriadoEm > 0 {
				instante := time.Unix(alteracao.Valor.CriadoEm, 0).UTC()
				evento.OcorridoEm = &instante
			}
			eventos = append(eventos, evento)
		}
	}
	return s.repositorio.Enfileirar(ctx, eventos)
}

func podeGerenciarIntegracao(papel domain.Role) bool {
	return papel == domain.RoleSuperAdmin || papel == domain.RoleGestor || papel == domain.RoleCorretorSolo
}

func validarAssinaturaMeta(corpo []byte, assinatura, segredo string) bool {
	assinatura = strings.TrimSpace(assinatura)
	if !strings.HasPrefix(assinatura, "sha256=") {
		return false
	}
	recebida, err := hex.DecodeString(strings.TrimPrefix(assinatura, "sha256="))
	if err != nil || len(recebida) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(segredo))
	_, _ = mac.Write(corpo)
	return hmac.Equal(recebida, mac.Sum(nil))
}

func compararSegredo(recebido, esperado string) bool {
	if len(recebido) != len(esperado) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(recebido), []byte(esperado)) == 1
}
