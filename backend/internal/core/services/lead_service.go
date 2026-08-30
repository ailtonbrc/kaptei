package services

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type leadService struct {
	leadRepo    ports.LeadRepository
	contaRepo   ports.ContaRepository
	usuarioRepo ports.UsuarioRepository
}

const versaoConsentimentoSite = "2026-08-05"

func NewLeadService(
	leadRepo ports.LeadRepository,
	contaRepo ports.ContaRepository,
	usuarioRepo ports.UsuarioRepository,
) ports.LeadService {
	return &leadService{
		leadRepo:    leadRepo,
		contaRepo:   contaRepo,
		usuarioRepo: usuarioRepo,
	}
}

func (s *leadService) ProcessarWebhook(ctx context.Context, token string, captura domain.CapturaLeadWebhook) error {
	token = strings.TrimSpace(token)
	var (
		conta *domain.ContaSaaS
		err   error
	)
	if formatoUUID.MatchString(token) {
		conta, err = s.contaRepo.GetByLeadToken(ctx, token)
	} else if len(token) == 64 && apenasHexadecimal(token) {
		conta, err = s.contaRepo.GetByLeadTokenHash(ctx, hashTokenRecuperacao(token))
	} else {
		return errors.New("token de integração inválido")
	}
	if err != nil || conta == nil {
		return errors.New("token de integração inválido ou conta não encontrada")
	}

	nome := strings.TrimSpace(captura.Nome)
	if nome == "" {
		nome = "Lead sem nome"
	}
	if len([]rune(nome)) > 120 {
		return errors.New("nome excede 120 caracteres")
	}
	email := strings.ToLower(strings.TrimSpace(captura.Email))
	telefone := strings.TrimSpace(captura.Telefone)
	origem := strings.TrimSpace(captura.Origem)
	mensagem := strings.TrimSpace(captura.Mensagem)
	if email == "" && telefone == "" {
		return errors.New("informe e-mail ou telefone")
	}
	if email != "" {
		endereco, err := mail.ParseAddress(email)
		if err != nil || !strings.EqualFold(endereco.Address, email) || len(email) > 254 {
			return errors.New("e-mail inválido")
		}
	}
	if telefone != "" && (len([]rune(telefone)) < 8 || len([]rune(telefone)) > 30) {
		return errors.New("telefone inválido")
	}
	if len([]rune(origem)) > 100 || len([]rune(mensagem)) > 2000 {
		return errors.New("origem ou mensagem excede o limite permitido")
	}

	lead := &domain.Lead{
		ContaID: conta.ID,
		Nome:    nome,
		Status:  domain.LeadStatusNovo,
	}
	if email != "" {
		lead.Email = &email
	}
	if telefone != "" {
		lead.Telefone = &telefone
	}
	if origem != "" {
		lead.Origem = &origem
	}
	if mensagem != "" {
		lead.Mensagem = &mensagem
	}

	return s.salvarComDistribuicao(ctx, conta, lead)
}

func apenasHexadecimal(valor string) bool {
	for _, caractere := range valor {
		if !((caractere >= '0' && caractere <= '9') || (caractere >= 'a' && caractere <= 'f') || (caractere >= 'A' && caractere <= 'F')) {
			return false
		}
	}
	return true
}

func (s *leadService) CaptarSite(ctx context.Context, contaID string, captura domain.CapturaLeadPublico, imovelID *string) error {
	conta, err := s.contaRepo.GetByID(ctx, contaID)
	if err != nil || conta == nil {
		return errors.New("site não encontrado")
	}
	if !captura.ConsentimentoLGPD {
		return errors.New("consentimento para contato é obrigatório")
	}
	chaveIdempotencia := strings.TrimSpace(captura.ChaveIdempotencia)
	if !formatoUUID.MatchString(chaveIdempotencia) {
		return errors.New("identificador da solicitação inválido")
	}
	nome := strings.TrimSpace(captura.Nome)
	if len(nome) < 2 || len(nome) > 120 {
		return errors.New("nome deve ter entre 2 e 120 caracteres")
	}
	captura.Email = textoOpcional(captura.Email)
	captura.Telefone = textoOpcional(captura.Telefone)
	if captura.Email == nil && captura.Telefone == nil {
		return errors.New("informe e-mail ou telefone")
	}
	if captura.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*captura.Email))
		if len(email) > 254 {
			return errors.New("e-mail inválido")
		}
		endereco, err := mail.ParseAddress(email)
		if err != nil || !strings.EqualFold(endereco.Address, email) {
			return errors.New("e-mail inválido")
		}
		captura.Email = &email
	}
	if captura.Telefone != nil {
		telefone := strings.TrimSpace(*captura.Telefone)
		if len(telefone) < 8 || len(telefone) > 30 {
			return errors.New("telefone inválido")
		}
		captura.Telefone = &telefone
	}

	origem := "SITE"
	agora := time.Now().UTC()
	versaoConsentimento := versaoConsentimentoSite
	lead := &domain.Lead{
		ContaID: contaID, ImovelID: imovelID, Nome: nome, Email: captura.Email,
		Telefone: captura.Telefone, Origem: &origem, Mensagem: limitarTexto(captura.Mensagem, 2000),
		PaginaOrigem: limitarTexto(captura.PaginaOrigem, 500), UTMSource: limitarTexto(captura.UTMSource, 120),
		UTMMedium: limitarTexto(captura.UTMMedium, 120), UTMCampaign: limitarTexto(captura.UTMCampaign, 180),
		ConsentimentoLGPD: true, ConsentimentoEm: &agora, ConsentimentoVersao: &versaoConsentimento,
		ChaveIdempotencia: &chaveIdempotencia, Status: domain.LeadStatusNovo,
	}
	return s.salvarComDistribuicao(ctx, conta, lead)
}

func (s *leadService) CaptarIntegracao(ctx context.Context, contaID string, captura domain.CapturaLeadIntegracao) error {
	conta, err := s.contaRepo.GetByID(ctx, contaID)
	if err != nil || conta == nil {
		return errors.New("conta da integraÃ§Ã£o nÃ£o encontrada")
	}
	if !formatoUUID.MatchString(captura.ChaveIdempotencia) {
		return errors.New("identificador idempotente da integraÃ§Ã£o invÃ¡lido")
	}
	nome := strings.TrimSpace(captura.Nome)
	if nome == "" {
		nome = "Lead sem nome"
	}
	if len([]rune(nome)) > 120 {
		return errors.New("nome excede 120 caracteres")
	}
	email := strings.ToLower(strings.TrimSpace(captura.Email))
	telefone := strings.TrimSpace(captura.Telefone)
	if email == "" && telefone == "" {
		return errors.New("lead da integraÃ§Ã£o nÃ£o possui e-mail nem telefone")
	}
	if email != "" {
		endereco, erroEmail := mail.ParseAddress(email)
		if erroEmail != nil || !strings.EqualFold(endereco.Address, email) || len(email) > 254 {
			return errors.New("e-mail recebido da integraÃ§Ã£o Ã© invÃ¡lido")
		}
	}
	if telefone != "" && (len([]rune(telefone)) < 8 || len([]rune(telefone)) > 30) {
		return errors.New("telefone recebido da integraÃ§Ã£o Ã© invÃ¡lido")
	}
	origem := strings.ToUpper(strings.TrimSpace(captura.Origem))
	if origem == "" || len([]rune(origem)) > 100 {
		return errors.New("origem da integraÃ§Ã£o invÃ¡lida")
	}
	mensagem := strings.TrimSpace(captura.Mensagem)
	chave := captura.ChaveIdempotencia
	lead := &domain.Lead{
		ContaID: contaID, ImovelID: captura.ImovelID, Nome: nome, Origem: &origem, Status: domain.LeadStatusNovo,
		ChaveIdempotencia: &chave,
	}
	if email != "" {
		lead.Email = &email
	}
	if telefone != "" {
		lead.Telefone = &telefone
	}
	if mensagem != "" {
		lead.Mensagem = limitarTexto(&mensagem, 2000)
	}
	return s.salvarComDistribuicao(ctx, conta, lead)
}

func textoOpcional(valor *string) *string {
	if valor == nil {
		return nil
	}
	texto := strings.TrimSpace(*valor)
	if texto == "" {
		return nil
	}
	return &texto
}

func (s *leadService) salvarComDistribuicao(ctx context.Context, conta *domain.ContaSaaS, lead *domain.Lead) error {
	if conta.TipoConta == "CORRETOR_SOLO" {
		usuarios, err := s.usuarioRepo.ListByContaID(ctx, conta.ID)
		if err != nil {
			return err
		}
		for _, usuario := range usuarios {
			if strings.EqualFold(usuario.Status, "ATIVO") && usuario.Papel == domain.RoleCorretorSolo {
				lead.UsuarioID = &usuario.ID
				break
			}
		}
	} else if conta.TipoConta == "IMOBILIARIA" {
		if conta.LeadEstrategia == "ROLETA" {
			return s.leadRepo.CreateDistribuido(ctx, lead)
		} else {
			// CAIXA_ENTRADA: Deixa o UsuarioID nulo para ir para a caixa central
			lead.UsuarioID = nil
		}
	}

	return s.leadRepo.Create(ctx, lead)
}

func limitarTexto(valor *string, limite int) *string {
	if valor == nil {
		return nil
	}
	texto := strings.TrimSpace(*valor)
	if texto == "" {
		return nil
	}
	runas := []rune(texto)
	if len(runas) > limite {
		texto = string(runas[:limite])
	}
	return &texto
}

func (s *leadService) List(ctx context.Context, contaID, usuarioID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Lead], error) {
	filtro = normalizarFiltroPaginacao(filtro)
	if papel != domain.RoleSuperAdmin && papel != domain.RoleGestor {
		filtro.UsuarioID = &usuarioID
	}
	return s.leadRepo.ListByContaID(ctx, contaID, filtro)
}

func (s *leadService) Atribuir(ctx context.Context, leadID, contaID, usuarioAtorID, usuarioDestinoID string, papel domain.Role) error {
	if papel != domain.RoleGestor && papel != domain.RoleSuperAdmin {
		return errors.New("apenas gestores podem atribuir leads")
	}
	usuarioDestino, err := s.usuarioRepo.GetByID(ctx, usuarioDestinoID)
	if err != nil || usuarioDestino == nil || usuarioDestino.ContaID != contaID {
		return errors.New("corretor de destino não pertence a esta conta")
	}
	if !strings.EqualFold(usuarioDestino.Status, "ATIVO") {
		return errors.New("corretor de destino está inativo")
	}
	if usuarioDestino.Papel != domain.RoleCorretorEquipe && usuarioDestino.Papel != domain.RoleCorretorSolo && usuarioDestino.Papel != domain.RoleGestor {
		return errors.New("usuário de destino não pode receber leads")
	}

	lead, err := s.leadRepo.GetByID(ctx, leadID, contaID)
	if err != nil || lead == nil {
		return errors.New("lead não encontrado")
	}

	lead.UsuarioID = &usuarioDestinoID
	lead.Status = domain.LeadStatusEmAtendimento

	return s.leadRepo.Update(ctx, lead)
}

func (s *leadService) Qualificar(ctx context.Context, leadID, contaID, usuarioAtorID string, papel domain.Role) error {
	if leadID == "" || contaID == "" {
		return errors.New("lead e conta são obrigatórios")
	}
	if err := s.validarAcessoAoLead(ctx, leadID, contaID, usuarioAtorID, papel); err != nil {
		return err
	}
	return s.leadRepo.Qualificar(ctx, leadID, contaID)
}

func (s *leadService) Descartar(ctx context.Context, leadID, contaID, usuarioAtorID, motivo string, papel domain.Role) error {
	lead, err := s.leadRepo.GetByID(ctx, leadID, contaID)
	if err != nil || lead == nil {
		return errors.New("lead não encontrado")
	}
	if papel != domain.RoleGestor && papel != domain.RoleSuperAdmin && (lead.UsuarioID == nil || *lead.UsuarioID != usuarioAtorID) {
		return errors.New("sem permissão para alterar este lead")
	}
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return errors.New("motivo do descarte é obrigatório")
	}
	if len([]rune(motivo)) > 500 {
		return errors.New("motivo do descarte excede 500 caracteres")
	}

	lead.Status = domain.LeadStatusDescartado
	lead.MotivoDescarte = &motivo

	return s.leadRepo.Update(ctx, lead)
}

func (s *leadService) validarAcessoAoLead(ctx context.Context, leadID, contaID, usuarioAtorID string, papel domain.Role) error {
	if papel == domain.RoleGestor || papel == domain.RoleSuperAdmin {
		return nil
	}
	lead, err := s.leadRepo.GetByID(ctx, leadID, contaID)
	if err != nil || lead == nil {
		return errors.New("lead não encontrado")
	}
	if lead.UsuarioID == nil || *lead.UsuarioID != usuarioAtorID {
		return errors.New("sem permissão para alterar este lead")
	}
	return nil
}
