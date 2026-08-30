package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type usuarioService struct{ repositorio ports.UsuarioRepository }

var siglaEstadoValida = regexp.MustCompile(`^[A-Za-z]{2}$`)

func NewUsuarioService(repositorio ports.UsuarioRepository) ports.UsuarioService {
	return &usuarioService{repositorio: repositorio}
}

func (s *usuarioService) AtualizarPerfil(ctx context.Context, usuarioID string, dados domain.AtualizacaoPerfil) (*domain.Usuario, error) {
	usuario, err := s.repositorio.GetByID(ctx, usuarioID)
	if err != nil || usuario == nil {
		return nil, errors.New("usuário não encontrado")
	}
	dados.NomeCompleto = strings.TrimSpace(dados.NomeCompleto)
	if utf8.RuneCountInString(dados.NomeCompleto) < 2 || utf8.RuneCountInString(dados.NomeCompleto) > 120 {
		return nil, errors.New("nome deve ter entre 2 e 120 caracteres")
	}
	campos := []struct {
		valor  **string
		limite int
		nome   string
	}{
		{&dados.Telefone, 30, "telefone"}, {&dados.NumeroWhatsapp, 30, "WhatsApp"}, {&dados.CPF, 20, "CPF"},
		{&dados.RG, 30, "RG"}, {&dados.RGEstado, 2, "estado do RG"}, {&dados.RGOrgaoExpedidor, 20, "órgão expedidor"},
		{&dados.Nacionalidade, 80, "nacionalidade"}, {&dados.EstadoCivil, 50, "estado civil"},
		{&dados.Creci, 30, "CRECI"}, {&dados.CreciEstado, 2, "estado do CRECI"}, {&dados.CEP, 20, "CEP"},
		{&dados.Logradouro, 255, "logradouro"}, {&dados.Numero, 50, "número"}, {&dados.Complemento, 150, "complemento"},
		{&dados.Bairro, 150, "bairro"}, {&dados.Cidade, 150, "cidade"}, {&dados.Estado, 2, "estado"},
	}
	for _, campo := range campos {
		*campo.valor = textoOpcional(*campo.valor)
		if *campo.valor != nil && utf8.RuneCountInString(**campo.valor) > campo.limite {
			return nil, errors.New(campo.nome + " excede o limite permitido")
		}
	}
	if dados.Telefone != nil && utf8.RuneCountInString(*dados.Telefone) < 8 {
		return nil, errors.New("telefone inválido")
	}
	if dados.NumeroWhatsapp != nil && utf8.RuneCountInString(*dados.NumeroWhatsapp) < 8 {
		return nil, errors.New("WhatsApp inválido")
	}
	if dados.CPF != nil && !cpfValido(*dados.CPF) {
		return nil, errors.New("CPF inválido")
	}
	for _, estado := range []*string{dados.RGEstado, dados.CreciEstado, dados.Estado} {
		if estado != nil && !siglaEstadoValida.MatchString(*estado) {
			return nil, errors.New("sigla de estado inválida")
		}
		if estado != nil {
			valor := strings.ToUpper(*estado)
			*estado = valor
		}
	}

	usuario.NomeCompleto = dados.NomeCompleto
	usuario.Telefone = dados.Telefone
	usuario.NumeroWhatsapp = dados.NumeroWhatsapp
	usuario.CPF = dados.CPF
	usuario.RG = dados.RG
	usuario.RGEstado = dados.RGEstado
	usuario.RGOrgaoExpedidor = dados.RGOrgaoExpedidor
	usuario.Nacionalidade = dados.Nacionalidade
	usuario.EstadoCivil = dados.EstadoCivil
	usuario.Creci = dados.Creci
	usuario.CreciEstado = dados.CreciEstado
	usuario.CEP = dados.CEP
	usuario.Logradouro = dados.Logradouro
	usuario.Numero = dados.Numero
	usuario.Complemento = dados.Complemento
	usuario.Bairro = dados.Bairro
	usuario.Cidade = dados.Cidade
	usuario.Estado = dados.Estado
	if err := s.repositorio.Update(ctx, usuario); err != nil {
		return nil, errors.New("não foi possível atualizar o perfil")
	}
	return usuario, nil
}

func (s *usuarioService) AlterarSenha(ctx context.Context, usuarioID, senhaAtual, novaSenha string) error {
	if senhaAtual == "" || len(novaSenha) < 6 || len([]byte(novaSenha)) > 72 {
		return errors.New("informe a senha atual e uma nova senha entre 6 e 72 bytes")
	}
	usuario, err := s.repositorio.GetByID(ctx, usuarioID)
	if err != nil || usuario == nil || usuario.SenhaHash == nil {
		return errors.New("não foi possível validar a senha atual")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*usuario.SenhaHash), []byte(senhaAtual)); err != nil {
		return errors.New("senha atual incorreta")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("não foi possível proteger a nova senha")
	}
	if err := s.repositorio.AtualizarSenha(ctx, usuario.ID, string(hash)); err != nil {
		return errors.New("não foi possível atualizar a senha")
	}
	return nil
}
