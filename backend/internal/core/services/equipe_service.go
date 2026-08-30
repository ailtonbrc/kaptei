package services

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type equipeService struct {
	usuarios        ports.UsuarioRepository
	convites        ports.ConviteEquipeRepository
	contas          ports.ContaRepository
	planos          ports.PlanoRepository
	preparadorEmail ports.PreparadorEmailOutbox
	urlPublica      string
}

func NewEquipeService(
	usuarios ports.UsuarioRepository,
	convites ports.ConviteEquipeRepository,
	contas ports.ContaRepository,
	planos ports.PlanoRepository,
	preparadorEmail ports.PreparadorEmailOutbox,
	urlPublica string,
) ports.EquipeService {
	return &equipeService{usuarios: usuarios, convites: convites, contas: contas, planos: planos, preparadorEmail: preparadorEmail, urlPublica: strings.TrimRight(urlPublica, "/")}
}

func (s *equipeService) Listar(ctx context.Context, contaID string, papelAtor domain.Role) ([]*domain.Usuario, []*domain.ConviteEquipe, error) {
	if papelAtor != domain.RoleGestor {
		return nil, nil, errors.New("apenas gestores podem administrar a equipe")
	}
	usuarios, err := s.usuarios.ListByContaID(ctx, contaID)
	if err != nil {
		return nil, nil, err
	}
	convites, err := s.convites.ListarPendentes(ctx, contaID)
	return usuarios, convites, err
}

func (s *equipeService) Convidar(ctx context.Context, contaID, usuarioAtorID, email string, papelAtor domain.Role) error {
	if papelAtor != domain.RoleGestor {
		return errors.New("apenas gestores podem convidar corretores")
	}
	if s.urlPublica == "" {
		return errors.New("URL pública da aplicação não configurada")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	endereco, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(endereco.Address, email) || len(email) > 254 {
		return errors.New("e-mail inválido")
	}
	if existente, err := s.usuarios.GetByEmail(ctx, email); err != nil {
		return errors.New("não foi possível validar o e-mail")
	} else if existente != nil {
		return errors.New("este e-mail já pertence a um usuário")
	}
	conta, plano, err := s.carregarContaEPlano(ctx, contaID)
	if err != nil {
		return err
	}
	if conta.TipoConta != "IMOBILIARIA" {
		return errors.New("gestão de equipe está disponível apenas para imobiliárias")
	}
	token, err := generateSecureToken()
	if err != nil {
		return errors.New("não foi possível gerar o convite")
	}
	convite := &domain.ConviteEquipe{
		ContaID: contaID, Email: email, Papel: domain.RoleCorretorEquipe,
		TokenHash: hashTokenRecuperacao(token), ConvidadoPor: usuarioAtorID,
		ExpiraEm: time.Now().UTC().Add(72 * time.Hour),
	}
	link := fmt.Sprintf("%s/aceitar-convite?token=%s", s.urlPublica, url.QueryEscape(token))
	corpo := fmt.Sprintf(`<h2>Convite para a equipe Kaptei</h2><p>Você foi convidado para integrar uma imobiliária no Kaptei.</p><p><a href="%s">Aceitar convite e criar senha</a></p><p>Este link expira em 72 horas e só pode ser usado uma vez.</p>`, html.EscapeString(link))
	evento, err := s.preparadorEmail.PrepararEmail(
		&contaID,
		"convite-equipe:"+convite.TokenHash,
		email,
		"Convite para a equipe - Kaptei",
		corpo,
	)
	if err != nil {
		return err
	}
	return s.convites.Criar(ctx, convite, plano.LimiteCorretores, evento)
}

func (s *equipeService) CancelarConvite(ctx context.Context, conviteID, contaID string, papelAtor domain.Role) error {
	if papelAtor != domain.RoleGestor {
		return errors.New("apenas gestores podem cancelar convites")
	}
	if strings.TrimSpace(conviteID) == "" {
		return errors.New("convite não informado")
	}
	return s.convites.Cancelar(ctx, conviteID, contaID)
}

func (s *equipeService) AtualizarStatus(ctx context.Context, usuarioID, contaID, usuarioAtorID, status string, papelAtor domain.Role) error {
	if papelAtor != domain.RoleGestor {
		return errors.New("apenas gestores podem alterar a equipe")
	}
	if usuarioID == usuarioAtorID {
		return errors.New("o gestor não pode alterar o próprio status")
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "ATIVO" && status != "INATIVO" {
		return errors.New("status deve ser ATIVO ou INATIVO")
	}
	_, plano, err := s.carregarContaEPlano(ctx, contaID)
	if err != nil {
		return err
	}
	return s.usuarios.AtualizarStatusEquipe(ctx, usuarioID, contaID, status, plano.LimiteCorretores)
}

func (s *equipeService) AceitarConvite(ctx context.Context, token, nome, senha string) error {
	token = strings.TrimSpace(token)
	nome = strings.TrimSpace(nome)
	if len(token) != 64 {
		return errors.New("convite inválido")
	}
	if utf8.RuneCountInString(nome) < 2 || utf8.RuneCountInString(nome) > 120 {
		return errors.New("nome deve ter entre 2 e 120 caracteres")
	}
	if len(senha) < 6 || len([]byte(senha)) > 72 {
		return errors.New("a senha deve ter entre 6 e 72 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("não foi possível proteger a senha")
	}
	_, err = s.convites.Aceitar(ctx, hashTokenRecuperacao(token), nome, string(hash))
	if err != nil {
		return errors.New("convite inválido, expirado, sem vaga disponível ou já utilizado")
	}
	return nil
}

func (s *equipeService) carregarContaEPlano(ctx context.Context, contaID string) (*domain.ContaSaaS, *domain.Plano, error) {
	conta, err := s.contas.GetByID(ctx, contaID)
	if err != nil || conta == nil {
		return nil, nil, errors.New("conta não encontrada")
	}
	plano, err := s.planos.GetByCodigo(ctx, conta.Plano)
	if err != nil || plano == nil {
		return nil, nil, errors.New("plano da conta não encontrado")
	}
	return conta, plano, nil
}
