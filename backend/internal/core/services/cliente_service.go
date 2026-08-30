package services

import (
	"context"
	"errors"
	"math"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type clienteService struct {
	repo        ports.ClienteRepository
	usuarioRepo ports.UsuarioRepository
}

func NewClienteService(repo ports.ClienteRepository, usuarioRepo ports.UsuarioRepository) ports.ClienteService {
	return &clienteService{repo: repo, usuarioRepo: usuarioRepo}
}

func (s *clienteService) Create(ctx context.Context, cliente *domain.Cliente, usuarioAtorID string, papel domain.Role) error {
	if err := validarCliente(cliente); err != nil {
		return err
	}
	if cliente.StatusFunil == "" {
		cliente.StatusFunil = "NOVO"
	}

	if !podeGerenciarClientes(papel) {
		cliente.CorretorID = &usuarioAtorID
	} else if cliente.CorretorID != nil {
		if err := s.validarCorretor(ctx, cliente.ContaID, *cliente.CorretorID); err != nil {
			return err
		}
	} else {
		cliente.CorretorID = s.proximoCorretor(ctx, cliente.ContaID)
	}

	return s.repo.Create(ctx, cliente)
}

func (s *clienteService) GetByID(ctx context.Context, id, contaID, usuarioAtorID string, papel domain.Role) (*domain.Cliente, error) {
	if id == "" || contaID == "" {
		return nil, errors.New("cliente e conta são obrigatórios")
	}
	cliente, err := s.repo.GetByID(ctx, id, contaID)
	if err != nil || cliente == nil {
		return cliente, err
	}
	if !podeGerenciarClientes(papel) && (cliente.CorretorID == nil || *cliente.CorretorID != usuarioAtorID) {
		return nil, errors.New("sem permissão para acessar este cliente")
	}
	return cliente, nil
}

func (s *clienteService) List(ctx context.Context, contaID, usuarioAtorID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Cliente], error) {
	if contaID == "" {
		return nil, errors.New("conta é obrigatória")
	}
	filtro = normalizarFiltroPaginacao(filtro)
	if !podeGerenciarClientes(papel) {
		filtro.UsuarioID = &usuarioAtorID
	}
	return s.repo.ListByContaID(ctx, contaID, filtro)
}

func (s *clienteService) Update(ctx context.Context, cliente *domain.Cliente, usuarioAtorID string, papel domain.Role) error {
	if cliente == nil || cliente.ID == "" {
		return errors.New("cliente é obrigatório")
	}
	if err := validarCliente(cliente); err != nil {
		return err
	}
	existente, err := s.repo.GetByID(ctx, cliente.ID, cliente.ContaID)
	if err != nil || existente == nil {
		return errors.New("cliente não encontrado")
	}
	if !podeGerenciarClientes(papel) {
		if existente.CorretorID == nil || *existente.CorretorID != usuarioAtorID {
			return errors.New("sem permissão para alterar este cliente")
		}
		cliente.CorretorID = &usuarioAtorID
	} else if cliente.CorretorID != nil {
		if err := s.validarCorretor(ctx, cliente.ContaID, *cliente.CorretorID); err != nil {
			return err
		}
	}
	return s.repo.Update(ctx, cliente)
}

func (s *clienteService) Delete(ctx context.Context, id, contaID, usuarioAtorID string, papel domain.Role) error {
	if id == "" || contaID == "" {
		return errors.New("cliente e conta são obrigatórios")
	}
	cliente, err := s.repo.GetByID(ctx, id, contaID)
	if err != nil || cliente == nil {
		return errors.New("cliente não encontrado")
	}
	if !podeGerenciarClientes(papel) && (cliente.CorretorID == nil || *cliente.CorretorID != usuarioAtorID) {
		return errors.New("sem permissão para excluir este cliente")
	}
	return s.repo.Delete(ctx, id, contaID)
}

func podeGerenciarClientes(papel domain.Role) bool {
	return papel == domain.RoleGestor || papel == domain.RoleSuperAdmin
}

func (s *clienteService) proximoCorretor(ctx context.Context, contaID string) *string {
	usuarios, err := s.usuarioRepo.ListByContaID(ctx, contaID)
	if err != nil {
		return nil
	}

	ativos := make([]string, 0, len(usuarios))
	for _, usuario := range usuarios {
		if strings.EqualFold(usuario.Status, "ATIVO") &&
			(usuario.Papel == domain.RoleCorretorEquipe || usuario.Papel == domain.RoleCorretorSolo) {
			ativos = append(ativos, usuario.ID)
		}
	}
	if len(ativos) == 0 {
		return nil
	}

	ultimo, _ := s.repo.GetUltimoCorretorAtribuido(ctx, contaID)
	proximoIndice := 0
	if ultimo != nil {
		for indice, id := range ativos {
			if id == *ultimo {
				proximoIndice = (indice + 1) % len(ativos)
				break
			}
		}
	}
	return &ativos[proximoIndice]
}

func (s *clienteService) validarCorretor(ctx context.Context, contaID, corretorID string) error {
	corretor, err := s.usuarioRepo.GetByID(ctx, corretorID)
	if err != nil || corretor == nil || corretor.ContaID != contaID {
		return errors.New("corretor não pertence a esta conta")
	}
	if !strings.EqualFold(corretor.Status, "ATIVO") {
		return errors.New("corretor está inativo")
	}
	if corretor.Papel != domain.RoleCorretorEquipe && corretor.Papel != domain.RoleCorretorSolo && corretor.Papel != domain.RoleGestor {
		return errors.New("usuário não pode ser responsável por clientes")
	}
	return nil
}

func validarCliente(cliente *domain.Cliente) error {
	if cliente == nil || cliente.ContaID == "" {
		return errors.New("conta é obrigatória")
	}
	cliente.Nome = strings.TrimSpace(cliente.Nome)
	if utf8.RuneCountInString(cliente.Nome) < 2 || utf8.RuneCountInString(cliente.Nome) > 255 {
		return errors.New("nome do cliente deve ter entre 2 e 255 caracteres")
	}
	cliente.Email = textoOpcional(cliente.Email)
	cliente.Telefone = textoOpcional(cliente.Telefone)
	cliente.CPF = textoOpcional(cliente.CPF)
	cliente.DataNascimento = textoOpcional(cliente.DataNascimento)
	cliente.EstadoCivil = textoOpcional(cliente.EstadoCivil)
	cliente.Origem = textoOpcional(cliente.Origem)
	cliente.InteresseTipo = textoOpcional(cliente.InteresseTipo)
	cliente.Notas = textoOpcional(cliente.Notas)
	cliente.Temperatura = textoOpcional(cliente.Temperatura)
	if cliente.Email != nil {
		email := strings.ToLower(*cliente.Email)
		endereco, err := mail.ParseAddress(email)
		if err != nil || len(email) > 254 || !strings.EqualFold(endereco.Address, email) {
			return errors.New("e-mail do cliente inválido")
		}
		cliente.Email = &email
	}
	if cliente.Telefone != nil && (utf8.RuneCountInString(*cliente.Telefone) < 8 || utf8.RuneCountInString(*cliente.Telefone) > 30) {
		return errors.New("telefone do cliente inválido")
	}
	if cliente.CPF != nil && !cpfValido(*cliente.CPF) {
		return errors.New("CPF do cliente inválido")
	}
	if cliente.DataNascimento != nil {
		data, err := time.Parse("2006-01-02", *cliente.DataNascimento)
		if err != nil || data.After(time.Now()) || data.Year() < 1900 {
			return errors.New("data de nascimento inválida")
		}
	}
	if cliente.EstadoCivil != nil && utf8.RuneCountInString(*cliente.EstadoCivil) > 50 {
		return errors.New("estado civil excede o limite permitido")
	}
	statusValidos := map[string]bool{"NOVO": true, "ATENDIMENTO": true, "VISITA": true, "PROPOSTA": true, "FECHADO": true, "PERDIDO": true}
	cliente.StatusFunil = strings.ToUpper(strings.TrimSpace(cliente.StatusFunil))
	if !statusValidos[cliente.StatusFunil] {
		return errors.New("status do funil inválido")
	}
	if cliente.Origem != nil {
		origem := strings.ToUpper(*cliente.Origem)
		origensValidas := map[string]bool{"SITE": true, "PORTAL": true, "WHATSAPP": true, "INDICACAO": true, "SOCIAL": true, "OUTROS": true}
		if !origensValidas[origem] {
			return errors.New("origem do cliente inválida")
		}
		cliente.Origem = &origem
	}
	if cliente.InteresseTipo != nil && *cliente.InteresseTipo != "Compra" && *cliente.InteresseTipo != "Locação" {
		return errors.New("tipo de interesse inválido")
	}
	if cliente.Notas != nil && utf8.RuneCountInString(*cliente.Notas) > 10_000 {
		return errors.New("notas excedem 10.000 caracteres")
	}
	if cliente.Temperatura != nil {
		temperatura := strings.ToUpper(*cliente.Temperatura)
		if temperatura != "FRIO" && temperatura != "MORNO" && temperatura != "QUENTE" {
			return errors.New("temperatura do cliente inválida")
		}
		cliente.Temperatura = &temperatura
	}
	if err := validarTagsCliente(cliente); err != nil {
		return err
	}
	if err := validarPreferenciasCliente(cliente.Preferencias); err != nil {
		return err
	}
	if err := validarFinanceiroCliente(cliente.Financeiro); err != nil {
		return err
	}
	if err := validarOrigemUTM(cliente.OrigemUTM); err != nil {
		return err
	}
	return nil
}

func validarTagsCliente(cliente *domain.Cliente) error {
	if len(cliente.Tags) > 20 {
		return errors.New("use no máximo 20 tags")
	}
	vistas := make(map[string]bool, len(cliente.Tags))
	normalizadas := make([]string, 0, len(cliente.Tags))
	for _, tag := range cliente.Tags {
		tag = strings.TrimSpace(tag)
		chave := strings.ToLower(tag)
		if tag == "" || utf8.RuneCountInString(tag) > 40 {
			return errors.New("cada tag deve possuir entre 1 e 40 caracteres")
		}
		if !vistas[chave] {
			vistas[chave] = true
			normalizadas = append(normalizadas, tag)
		}
	}
	cliente.Tags = normalizadas
	return nil
}

func validarPreferenciasCliente(preferencias *domain.ClientePreferencias) error {
	if preferencias == nil {
		return nil
	}
	if len(preferencias.TipoImovel) > 10 || len(preferencias.Bairros) > 30 {
		return errors.New("preferências excedem o limite permitido")
	}
	for _, valor := range append(append([]string{}, preferencias.TipoImovel...), preferencias.Bairros...) {
		if strings.TrimSpace(valor) == "" || utf8.RuneCountInString(valor) > 120 {
			return errors.New("item de preferência inválido")
		}
	}
	if preferencias.Finalidade != "" && preferencias.Finalidade != "Compra" && preferencias.Finalidade != "Locação" {
		return errors.New("finalidade de preferência inválida")
	}
	for _, valor := range []*float64{preferencias.OrcamentoMin, preferencias.OrcamentoMax} {
		if valor != nil && (math.IsNaN(*valor) || math.IsInf(*valor, 0) || *valor < 0 || *valor > 999_999_999_999) {
			return errors.New("orçamento inválido")
		}
	}
	if preferencias.OrcamentoMin != nil && preferencias.OrcamentoMax != nil && *preferencias.OrcamentoMin > *preferencias.OrcamentoMax {
		return errors.New("orçamento mínimo não pode superar o máximo")
	}
	if preferencias.QuartosMin != nil && (*preferencias.QuartosMin < 0 || *preferencias.QuartosMin > 100) {
		return errors.New("quantidade mínima de quartos inválida")
	}
	return nil
}

func validarFinanceiroCliente(financeiro *domain.ClienteFinanceiro) error {
	if financeiro == nil {
		return nil
	}
	if financeiro.RendaMensal != nil && (math.IsNaN(*financeiro.RendaMensal) || math.IsInf(*financeiro.RendaMensal, 0) || *financeiro.RendaMensal < 0 || *financeiro.RendaMensal > 999_999_999_999) {
		return errors.New("renda mensal inválida")
	}
	if financeiro.PrecisaFinanciamento != nil && *financeiro.PrecisaFinanciamento != "" && *financeiro.PrecisaFinanciamento != "Sim" && *financeiro.PrecisaFinanciamento != "Nao" && *financeiro.PrecisaFinanciamento != "JaAprovado" {
		return errors.New("situação de financiamento inválida")
	}
	if financeiro.FormaPagamento != nil && *financeiro.FormaPagamento != "" && *financeiro.FormaPagamento != "AVista" && *financeiro.FormaPagamento != "Financiamento" && *financeiro.FormaPagamento != "Consorcio" && *financeiro.FormaPagamento != "Parcelado" {
		return errors.New("forma de pagamento inválida")
	}
	return nil
}

var formatoUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
var caracteresNaoDigitos = regexp.MustCompile(`\D`)

func validarOrigemUTM(origem *domain.ClienteOrigemUTM) error {
	if origem == nil {
		return nil
	}
	origem.Canal = strings.TrimSpace(origem.Canal)
	origem.Campanha = strings.TrimSpace(origem.Campanha)
	origem.ImovelOrigemID = strings.TrimSpace(origem.ImovelOrigemID)
	if utf8.RuneCountInString(origem.Canal) > 120 || utf8.RuneCountInString(origem.Campanha) > 180 {
		return errors.New("atribuição UTM excede o limite permitido")
	}
	if origem.ImovelOrigemID != "" && !formatoUUID.MatchString(origem.ImovelOrigemID) {
		return errors.New("identificador do imóvel de origem inválido")
	}
	return nil
}

func cpfValido(valor string) bool {
	digitos := caracteresNaoDigitos.ReplaceAllString(valor, "")
	if len(digitos) != 11 || strings.Count(digitos, string(digitos[0])) == 11 {
		return false
	}
	calcular := func(tamanho, peso int) byte {
		soma := 0
		for indice := 0; indice < tamanho; indice++ {
			soma += int(digitos[indice]-'0') * (peso - indice)
		}
		resto := (soma * 10) % 11
		if resto == 10 {
			resto = 0
		}
		return byte(resto) + '0'
	}
	return digitos[9] == calcular(9, 10) && digitos[10] == calcular(10, 11)
}
