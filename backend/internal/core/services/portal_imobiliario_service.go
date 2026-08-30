package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

var (
	ErrPortalNaoAutorizado = errors.New("webhook do portal não autorizado")
	ErrLeadPortalInvalido  = errors.New("lead do portal inválido")
	ErrPortalIndisponivel  = errors.New("webhook do portal indisponível")
)

type portalImobiliarioService struct {
	repositorio    ports.PortalImobiliarioRepository
	gerador        ports.GeradorFeedPortal
	origemPublica  string
	leads          ports.LeadService
	segredoWebhook string
}

func NewPortalImobiliarioService(repositorio ports.PortalImobiliarioRepository, gerador ports.GeradorFeedPortal, leads ports.LeadService, origemPublica, segredoWebhook string) ports.PortalImobiliarioService {
	return &portalImobiliarioService{repositorio: repositorio, gerador: gerador, leads: leads, origemPublica: strings.TrimRight(origemPublica, "/"), segredoWebhook: strings.TrimSpace(segredoWebhook)}
}

func (s *portalImobiliarioService) ObterConfiguracao(ctx context.Context, contaID string, papel domain.Role) (*domain.ConfiguracaoPortal, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissão para administrar portais")
	}
	return s.repositorio.ObterConfiguracao(ctx, contaID, domain.PortalGrupoOLX)
}

func (s *portalImobiliarioService) SalvarConfiguracao(ctx context.Context, contaID, usuarioID string, papel domain.Role, configuracao domain.ConfiguracaoPortal) (*domain.ConfiguracaoPortal, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissão para administrar portais")
	}
	configuracao.ContaID = contaID
	configuracao.Portal = domain.PortalGrupoOLX
	configuracao.NomeContato = strings.TrimSpace(configuracao.NomeContato)
	configuracao.EmailContato = strings.ToLower(strings.TrimSpace(configuracao.EmailContato))
	configuracao.TelefoneContato = strings.TrimSpace(configuracao.TelefoneContato)
	configuracao.ExibicaoEndereco = strings.ToUpper(strings.TrimSpace(configuracao.ExibicaoEndereco))
	if err := validarConfiguracaoPortal(configuracao); err != nil {
		return nil, err
	}
	atual, err := s.repositorio.ObterConfiguracao(ctx, contaID, domain.PortalGrupoOLX)
	if err != nil {
		return nil, err
	}
	configuracao.TokenFeedPrefixo = atual.TokenFeedPrefixo
	if configuracao.Ativa {
		if configuracao.TokenFeedPrefixo == nil {
			return nil, errors.New("gere a URL protegida do feed antes de ativar")
		}
		dados, err := s.repositorio.ObterDadosFeed(ctx, contaID, domain.PortalGrupoOLX)
		if err != nil {
			return nil, err
		}
		dados.Configuracao = configuracao
		if diagnostico := s.gerador.Validar(dados); !diagnostico.Valido {
			return nil, fmt.Errorf("corrija o feed antes de ativar: %s", resumirErrosPortal(diagnostico))
		}
	}
	if err := s.repositorio.SalvarConfiguracao(ctx, &configuracao, usuarioID); err != nil {
		return nil, err
	}
	return s.repositorio.ObterConfiguracao(ctx, contaID, domain.PortalGrupoOLX)
}

func (s *portalImobiliarioService) RotacionarToken(ctx context.Context, contaID, usuarioID string, papel domain.Role) (*domain.CredencialFeedPortal, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissão para administrar portais")
	}
	bytesToken := make([]byte, 32)
	if _, err := rand.Read(bytesToken); err != nil {
		return nil, fmt.Errorf("gerar token do feed: %w", err)
	}
	token := hex.EncodeToString(bytesToken)
	hash := hashTokenPortal(token)
	prefixo := token[:10]
	if err := s.repositorio.RotacionarToken(ctx, contaID, domain.PortalGrupoOLX, hash, prefixo, usuarioID); err != nil {
		return nil, err
	}
	basePortal := s.origemPublica + "/api"
	return &domain.CredencialFeedPortal{
		Token: token, URLFeed: basePortal + "/public/portais/grupo-olx/" + url.PathEscape(token) + "/vrsync.xml",
		URLWebhook: basePortal + "/webhooks/portais/grupo-olx/" + url.PathEscape(token) + "/leads",
	}, nil
}

func (s *portalImobiliarioService) ListarPublicacoes(ctx context.Context, contaID string, papel domain.Role) ([]domain.PublicacaoPortal, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissão para administrar portais")
	}
	publicacoes, err := s.repositorio.ListarPublicacoes(ctx, contaID, domain.PortalGrupoOLX)
	if err != nil {
		return nil, err
	}
	dados, err := s.repositorio.ObterDadosFeed(ctx, contaID, domain.PortalGrupoOLX)
	if err != nil {
		return nil, err
	}
	diagnostico := s.gerador.Validar(dados)
	errosPorImovel := make(map[string][]string, len(diagnostico.Publicacoes))
	for _, publicacao := range diagnostico.Publicacoes {
		errosPorImovel[publicacao.ImovelID] = publicacao.Erros
	}
	for indice := range publicacoes {
		if publicacoes[indice].Ativa {
			publicacoes[indice].Erros = errosPorImovel[publicacoes[indice].ImovelID]
		}
	}
	return publicacoes, nil
}

func (s *portalImobiliarioService) SalvarPublicacao(ctx context.Context, contaID, imovelID, usuarioID string, papel domain.Role, atualizacao domain.AtualizacaoPublicacaoPortal) error {
	if !podeGerenciarIntegracao(papel) {
		return errors.New("sem permissão para administrar portais")
	}
	if !formatoUUID.MatchString(imovelID) {
		return errors.New("identificador do imóvel inválido")
	}
	atualizacao.TipoPublicacao = strings.ToUpper(strings.TrimSpace(atualizacao.TipoPublicacao))
	if atualizacao.TipoPublicacao != "STANDARD" && atualizacao.TipoPublicacao != "PREMIUM" && atualizacao.TipoPublicacao != "SUPER_PREMIUM" {
		return errors.New("tipo de publicação inválido")
	}
	return s.repositorio.SalvarPublicacao(ctx, contaID, domain.PortalGrupoOLX, imovelID, usuarioID, atualizacao)
}

func (s *portalImobiliarioService) Diagnosticar(ctx context.Context, contaID string, papel domain.Role) (*domain.DiagnosticoFeedPortal, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissão para administrar portais")
	}
	dados, err := s.repositorio.ObterDadosFeed(ctx, contaID, domain.PortalGrupoOLX)
	if err != nil {
		return nil, err
	}
	return s.gerador.Validar(dados), nil
}

func (s *portalImobiliarioService) GerarFeedPublico(ctx context.Context, token string) ([]byte, error) {
	token = strings.TrimSpace(token)
	if len(token) != 64 {
		return nil, errors.New("feed não encontrado")
	}
	if _, err := hex.DecodeString(token); err != nil {
		return nil, errors.New("feed não encontrado")
	}
	contaID, err := s.repositorio.ObterContaPorToken(ctx, domain.PortalGrupoOLX, hashTokenPortal(token))
	if err != nil {
		return nil, err
	}
	dados, err := s.repositorio.ObterDadosFeed(ctx, contaID, domain.PortalGrupoOLX)
	if err != nil {
		return nil, err
	}
	return s.gerador.Gerar(dados, s.origemPublica, time.Now())
}

func (s *portalImobiliarioService) ReceberLead(ctx context.Context, token, autorizacao string, lead domain.LeadGrupoOLX) error {
	if s.segredoWebhook == "" || s.leads == nil {
		return ErrPortalIndisponivel
	}
	if !autorizacaoBasicaValida(autorizacao, s.segredoWebhook) {
		return ErrPortalNaoAutorizado
	}
	token = strings.TrimSpace(token)
	if len(token) != 64 {
		return ErrPortalNaoAutorizado
	}
	if _, err := hex.DecodeString(token); err != nil {
		return ErrPortalNaoAutorizado
	}
	contaID, err := s.repositorio.ObterContaPorToken(ctx, domain.PortalGrupoOLX, hashTokenPortal(token))
	if err != nil {
		return ErrPortalNaoAutorizado
	}

	lead.LeadOrigin = strings.TrimSpace(lead.LeadOrigin)
	lead.OriginLeadID = strings.TrimSpace(lead.OriginLeadID)
	if lead.LeadOrigin != "Grupo OLX" && lead.LeadOrigin != "MCMV_OLX" {
		return fmt.Errorf("%w: origem desconhecida", ErrLeadPortalInvalido)
	}
	if len(lead.OriginLeadID) < 1 || len(lead.OriginLeadID) > 255 {
		return fmt.Errorf("%w: originLeadId ausente ou excessivo", ErrLeadPortalInvalido)
	}
	if _, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(lead.Timestamp)); err != nil {
		return fmt.Errorf("%w: timestamp inválido", ErrLeadPortalInvalido)
	}
	nome := strings.TrimSpace(lead.Name)
	email := strings.ToLower(strings.TrimSpace(lead.Email))
	telefone := strings.TrimSpace(lead.DDD + lead.Phone)
	if telefone == "" {
		telefone = strings.TrimSpace(lead.PhoneNumber)
	}
	if utf8.RuneCountInString(nome) > 120 {
		return fmt.Errorf("%w: nome excede 120 caracteres", ErrLeadPortalInvalido)
	}
	if email == "" && telefone == "" {
		return fmt.Errorf("%w: informe e-mail ou telefone", ErrLeadPortalInvalido)
	}
	if email != "" {
		endereco, erroEmail := mail.ParseAddress(email)
		if erroEmail != nil || !strings.EqualFold(endereco.Address, email) || len(email) > 254 {
			return fmt.Errorf("%w: e-mail inválido", ErrLeadPortalInvalido)
		}
	}
	if telefone != "" && (utf8.RuneCountInString(telefone) < 8 || utf8.RuneCountInString(telefone) > 30) {
		return fmt.Errorf("%w: telefone inválido", ErrLeadPortalInvalido)
	}
	if lead.TransactionType != "" && lead.TransactionType != "SELL" && lead.TransactionType != "RENT" {
		return fmt.Errorf("%w: transação inválida", ErrLeadPortalInvalido)
	}
	if lead.Temperature != "" && lead.Temperature != "Baixa" && lead.Temperature != "Média" && lead.Temperature != "Alta" {
		return fmt.Errorf("%w: temperatura inválida", ErrLeadPortalInvalido)
	}
	if lead.ExtraData.LeadType != "" {
		canais := map[string]bool{"CLICK_SCHEDULE": true, "CLICK_WHATSAPP": true, "CONTACT_CHAT": true, "CONTACT_FORM": true, "PHONE_VIEW": true, "VISIT_REQUEST": true}
		if !canais[lead.ExtraData.LeadType] {
			return fmt.Errorf("%w: canal inválido", ErrLeadPortalInvalido)
		}
	}
	lead.Name = nome
	lead.Email = email

	var imovelID *string
	if lead.LeadOrigin == "Grupo OLX" {
		if !formatoUUID.MatchString(strings.TrimSpace(lead.ClientListingID)) {
			return fmt.Errorf("%w: clientListingId inválido", ErrLeadPortalInvalido)
		}
		imovelID, err = s.repositorio.ObterImovelDaConta(ctx, contaID, lead.ClientListingID)
		if err != nil {
			return err
		}
		if imovelID == nil {
			return fmt.Errorf("%w: imóvel não pertence ao inventário", ErrLeadPortalInvalido)
		}
	}
	detalhes := make([]string, 0, 5)
	if mensagem := strings.TrimSpace(lead.Message); mensagem != "" {
		detalhes = append(detalhes, mensagem)
	}
	if lead.Temperature != "" {
		detalhes = append(detalhes, "Temperatura: "+strings.TrimSpace(lead.Temperature))
	}
	if lead.TransactionType != "" {
		detalhes = append(detalhes, "Transação: "+strings.TrimSpace(lead.TransactionType))
	}
	if lead.ExtraData.LeadType != "" {
		detalhes = append(detalhes, "Canal: "+strings.TrimSpace(lead.ExtraData.LeadType))
	}
	if lead.ExtraData.LeadCerto {
		detalhes = append(detalhes, "LeadCerto: sim")
	}
	origem := "GRUPO_OLX"
	if lead.LeadOrigin == "MCMV_OLX" {
		origem = "MCMV_OLX"
	}
	return s.leads.CaptarIntegracao(ctx, contaID, domain.CapturaLeadIntegracao{
		ImovelID: imovelID, Nome: lead.Name, Email: lead.Email, Telefone: telefone, Origem: origem,
		Mensagem: strings.Join(detalhes, " | "), ChaveIdempotencia: uuidDeterministico("grupo-olx:" + lead.OriginLeadID),
	})
}

func autorizacaoBasicaValida(cabecalho, segredoEsperado string) bool {
	partes := strings.Fields(cabecalho)
	if len(partes) != 2 || !strings.EqualFold(partes[0], "Basic") {
		return false
	}
	decodificado, err := base64.StdEncoding.DecodeString(partes[1])
	if err != nil {
		return false
	}
	credenciais := strings.SplitN(string(decodificado), ":", 2)
	if len(credenciais) != 2 || credenciais[0] == "" {
		return false
	}
	recebido := sha256.Sum256([]byte(credenciais[1]))
	esperado := sha256.Sum256([]byte(segredoEsperado))
	return subtle.ConstantTimeCompare(recebido[:], esperado[:]) == 1
}

func validarConfiguracaoPortal(configuracao domain.ConfiguracaoPortal) error {
	if tamanho := utf8.RuneCountInString(configuracao.NomeContato); tamanho < 2 || tamanho > 120 {
		return errors.New("nome de contato deve ter entre 2 e 120 caracteres")
	}
	if _, err := mail.ParseAddress(configuracao.EmailContato); err != nil || len(configuracao.EmailContato) > 254 {
		return errors.New("e-mail de contato inválido")
	}
	if utf8.RuneCountInString(configuracao.TelefoneContato) > 32 || strings.ContainsAny(configuracao.TelefoneContato, "\r\n") {
		return errors.New("telefone de contato inválido")
	}
	if configuracao.ExibicaoEndereco != "BAIRRO" && configuracao.ExibicaoEndereco != "LOGRADOURO" && configuracao.ExibicaoEndereco != "COMPLETO" {
		return errors.New("exibição de endereço inválida")
	}
	return nil
}

func hashTokenPortal(token string) string {
	soma := sha256.Sum256([]byte(token))
	return hex.EncodeToString(soma[:])
}

func resumirErrosPortal(diagnostico *domain.DiagnosticoFeedPortal) string {
	partes := append([]string{}, diagnostico.ErrosGerais...)
	for _, publicacao := range diagnostico.Publicacoes {
		if len(publicacao.Erros) > 0 {
			partes = append(partes, publicacao.Titulo+": "+strings.Join(publicacao.Erros, ", "))
		}
	}
	return strings.Join(partes, "; ")
}
