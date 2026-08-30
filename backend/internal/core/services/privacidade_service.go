package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type privacidadeService struct {
	repo     ports.PrivacidadeRepository
	protetor ports.ProtetorSegredos
	agora    func() time.Time
}

type contatoTitular struct {
	Email    string `json:"email,omitempty"`
	Telefone string `json:"telefone,omitempty"`
}

func NewPrivacidadeService(repo ports.PrivacidadeRepository, protetor ports.ProtetorSegredos) ports.PrivacidadeService {
	return &privacidadeService{repo: repo, protetor: protetor, agora: time.Now}
}

func (s *privacidadeService) CriarPublica(ctx context.Context, slug string, nova domain.NovaSolicitacaoTitular) (string, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", errors.New("site não informado")
	}
	if err := normalizarSolicitacaoTitular(&nova); err != nil {
		return "", err
	}
	contaID, err := s.repo.ObterContaPorSlug(ctx, slug)
	if err != nil {
		return "", fmt.Errorf("localizar canal de privacidade: %w", err)
	}
	if contaID == "" {
		return "", errors.New("canal de privacidade indisponível")
	}
	protocolo, err := gerarProtocoloTitular(s.agora())
	if err != nil {
		return "", err
	}
	nomeProtegido, err := s.protetor.Proteger(nova.Nome)
	if err != nil {
		return "", fmt.Errorf("proteger nome do titular: %w", err)
	}
	contatoJSON, _ := json.Marshal(contatoTitular{Email: nova.Email, Telefone: nova.Telefone})
	contatoProtegido, err := s.protetor.Proteger(string(contatoJSON))
	if err != nil {
		return "", fmt.Errorf("proteger contato do titular: %w", err)
	}
	var detalhesProtegidos *string
	if nova.Detalhes != "" {
		protegido, err := s.protetor.Proteger(nova.Detalhes)
		if err != nil {
			return "", fmt.Errorf("proteger detalhes da solicitação: %w", err)
		}
		detalhesProtegidos = &protegido
	}
	solicitacao := &domain.SolicitacaoTitular{
		ContaID: contaID, Protocolo: protocolo, Tipo: nova.Tipo, Status: "RECEBIDA",
		PrazoRespostaEm: s.agora().AddDate(0, 0, 15), NomeProtegido: nomeProtegido,
		ContatoProtegido: contatoProtegido, DetalhesProtegidos: detalhesProtegidos,
		EmailHash: hashOpcional(nova.Email), TelefoneHash: hashOpcional(nova.Telefone),
	}
	if err := s.repo.Criar(ctx, solicitacao); err != nil {
		return "", err
	}
	return protocolo, nil
}

func (s *privacidadeService) Listar(ctx context.Context, contaID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.SolicitacaoTitular], error) {
	if !podeGerenciarPrivacidade(papel) {
		return nil, errors.New("sem permissão para gerenciar privacidade")
	}
	filtro = normalizarFiltroPaginacao(filtro)
	return s.repo.Listar(ctx, contaID, filtro)
}

func (s *privacidadeService) Obter(ctx context.Context, id, contaID string, papel domain.Role) (*domain.SolicitacaoTitular, error) {
	if !podeGerenciarPrivacidade(papel) {
		return nil, errors.New("sem permissão para gerenciar privacidade")
	}
	solicitacao, err := s.repo.Obter(ctx, id, contaID)
	if err != nil || solicitacao == nil {
		return solicitacao, err
	}
	if err := s.revelarSolicitacao(solicitacao); err != nil {
		return nil, err
	}
	return solicitacao, nil
}

func (s *privacidadeService) VerificarIdentidade(ctx context.Context, id, contaID, usuarioID string, papel domain.Role, metodo, evidencia string) error {
	if !podeGerenciarPrivacidade(papel) {
		return errors.New("sem permissão para verificar identidade")
	}
	metodo = strings.TrimSpace(metodo)
	evidencia = strings.TrimSpace(evidencia)
	if utf8.RuneCountInString(metodo) < 3 || utf8.RuneCountInString(metodo) > 80 {
		return errors.New("método de verificação deve ter entre 3 e 80 caracteres")
	}
	if utf8.RuneCountInString(evidencia) < 3 || utf8.RuneCountInString(evidencia) > 2_000 {
		return errors.New("evidência de verificação deve ter entre 3 e 2.000 caracteres")
	}
	protegida, err := s.protetor.Proteger(evidencia)
	if err != nil {
		return fmt.Errorf("proteger evidência de identidade: %w", err)
	}
	return s.repo.VerificarIdentidade(ctx, id, contaID, usuarioID, metodo, protegida)
}

func (s *privacidadeService) Decidir(ctx context.Context, id, contaID, usuarioID string, papel domain.Role, decisao domain.DecisaoSolicitacaoTitular) error {
	if !podeGerenciarPrivacidade(papel) {
		return errors.New("sem permissão para decidir solicitações")
	}
	decisao.FundamentoLegal = strings.TrimSpace(decisao.FundamentoLegal)
	decisao.Observacao = strings.TrimSpace(decisao.Observacao)
	if utf8.RuneCountInString(decisao.FundamentoLegal) < 5 || utf8.RuneCountInString(decisao.FundamentoLegal) > 2_000 {
		return errors.New("fundamento legal deve ter entre 5 e 2.000 caracteres")
	}
	if utf8.RuneCountInString(decisao.Observacao) > 2_000 {
		return errors.New("observação da decisão excede 2.000 caracteres")
	}
	valor := "REJEITADA"
	if decisao.Aprovada {
		valor = "APROVADA"
	}
	return s.repo.Decidir(ctx, id, contaID, usuarioID, valor, decisao.FundamentoLegal, decisao.Observacao)
}

func (s *privacidadeService) Exportar(ctx context.Context, id, contaID, usuarioID string, papel domain.Role) (*domain.ExportacaoTitular, error) {
	if !podeGerenciarPrivacidade(papel) {
		return nil, errors.New("sem permissão para exportar dados pessoais")
	}
	solicitacao, err := s.repo.Obter(ctx, id, contaID)
	if err != nil || solicitacao == nil {
		return nil, errors.New("solicitação não encontrada")
	}
	if solicitacao.Status != "APROVADA" || solicitacao.IdentidadeVerificadaEm == nil {
		return nil, errors.New("solicitação precisa estar aprovada e com identidade verificada")
	}
	if !tipoPermiteExportacao(solicitacao.Tipo) {
		return nil, errors.New("tipo da solicitação não permite exportação")
	}
	if err := s.revelarSolicitacao(solicitacao); err != nil {
		return nil, err
	}
	dados, err := s.repo.GerarDadosExportacao(ctx, contaID, solicitacao.Email, solicitacao.Telefone)
	if err != nil {
		return nil, err
	}
	if err := s.revelarMensagensExportadas(dados.Mensagens); err != nil {
		return nil, err
	}
	if err := s.repo.ConcluirExportacao(ctx, id, contaID, usuarioID); err != nil {
		return nil, err
	}
	return &domain.ExportacaoTitular{Protocolo: solicitacao.Protocolo, GeradaEm: s.agora(), Dados: map[string]any{
		"clientes": dados.Clientes, "leads": dados.Leads, "interacoes": dados.Interacoes,
		"agendamentos": dados.Agendamentos, "conversas_whatsapp": dados.Conversas, "mensagens_whatsapp": dados.Mensagens,
	}}, nil
}

func (s *privacidadeService) Executar(ctx context.Context, id, contaID, usuarioID string, papel domain.Role) error {
	if !podeGerenciarPrivacidade(papel) {
		return errors.New("sem permissão para executar direitos do titular")
	}
	solicitacao, err := s.repo.Obter(ctx, id, contaID)
	if err != nil || solicitacao == nil {
		return errors.New("solicitação não encontrada")
	}
	permitidos := map[domain.TipoSolicitacaoTitular]bool{domain.TipoExclusao: true, domain.TipoAnonimizacao: true, domain.TipoRevogacao: true, domain.TipoCorrecao: true, domain.TipoBloqueio: true}
	if !permitidos[solicitacao.Tipo] {
		return errors.New("tipo da solicitação exige exportação ou resposta administrativa")
	}
	if solicitacao.Status != "APROVADA" || solicitacao.IdentidadeVerificadaEm == nil {
		return errors.New("solicitação precisa estar aprovada e com identidade verificada")
	}
	if err := s.revelarSolicitacao(solicitacao); err != nil {
		return err
	}
	return s.repo.ExecutarDireito(ctx, id, contaID, usuarioID, solicitacao.Email, solicitacao.Telefone, solicitacao.Tipo)
}

func (s *privacidadeService) revelarSolicitacao(solicitacao *domain.SolicitacaoTitular) error {
	nome, err := s.protetor.Revelar(solicitacao.NomeProtegido)
	if err != nil {
		return fmt.Errorf("revelar nome do titular: %w", err)
	}
	contatoBruto, err := s.protetor.Revelar(solicitacao.ContatoProtegido)
	if err != nil {
		return fmt.Errorf("revelar contato do titular: %w", err)
	}
	var contato contatoTitular
	if err := json.Unmarshal([]byte(contatoBruto), &contato); err != nil {
		return fmt.Errorf("decodificar contato do titular: %w", err)
	}
	solicitacao.Nome, solicitacao.Email, solicitacao.Telefone = nome, contato.Email, contato.Telefone
	if solicitacao.DetalhesProtegidos != nil {
		detalhes, err := s.protetor.Revelar(*solicitacao.DetalhesProtegidos)
		if err != nil {
			return fmt.Errorf("revelar detalhes da solicitação: %w", err)
		}
		solicitacao.Detalhes = detalhes
	}
	return nil
}

func (s *privacidadeService) revelarMensagensExportadas(mensagens []map[string]any) error {
	for _, mensagem := range mensagens {
		protegido, ok := mensagem["conteudo_protegido"].(string)
		if !ok || protegido == "" {
			continue
		}
		revelado, err := s.protetor.Revelar(protegido)
		if err != nil {
			return fmt.Errorf("revelar mensagem da exportação: %w", err)
		}
		var conteudo any
		if json.Unmarshal([]byte(revelado), &conteudo) != nil {
			conteudo = revelado
		}
		delete(mensagem, "conteudo_protegido")
		mensagem["conteudo"] = conteudo
	}
	return nil
}

func normalizarSolicitacaoTitular(nova *domain.NovaSolicitacaoTitular) error {
	nova.Nome = strings.TrimSpace(nova.Nome)
	nova.Email = strings.ToLower(strings.TrimSpace(nova.Email))
	nova.Telefone = regexp.MustCompile(`\D`).ReplaceAllString(nova.Telefone, "")
	nova.Detalhes = strings.TrimSpace(nova.Detalhes)
	if utf8.RuneCountInString(nova.Nome) < 2 || utf8.RuneCountInString(nova.Nome) > 160 {
		return errors.New("nome deve ter entre 2 e 160 caracteres")
	}
	if nova.Email == "" && nova.Telefone == "" {
		return errors.New("informe e-mail ou telefone para identificação")
	}
	if nova.Email != "" {
		endereco, err := mail.ParseAddress(nova.Email)
		if err != nil || endereco.Address != nova.Email || len(nova.Email) > 254 {
			return errors.New("e-mail inválido")
		}
	}
	if nova.Telefone != "" && (len(nova.Telefone) < 8 || len(nova.Telefone) > 20) {
		return errors.New("telefone inválido")
	}
	if utf8.RuneCountInString(nova.Detalhes) > 4_000 {
		return errors.New("detalhes excedem 4.000 caracteres")
	}
	if !tipoSolicitacaoValido(nova.Tipo) {
		return errors.New("tipo de solicitação inválido")
	}
	return nil
}

func tipoSolicitacaoValido(tipo domain.TipoSolicitacaoTitular) bool {
	validos := map[domain.TipoSolicitacaoTitular]bool{
		domain.TipoConfirmacao: true, domain.TipoAcesso: true, domain.TipoCorrecao: true,
		domain.TipoAnonimizacao: true, domain.TipoBloqueio: true, domain.TipoExclusao: true,
		domain.TipoPortabilidade: true, domain.TipoRevogacao: true, domain.TipoInformacaoCompartilhamento: true,
	}
	return validos[tipo]
}

func tipoPermiteExportacao(tipo domain.TipoSolicitacaoTitular) bool {
	return tipo == domain.TipoConfirmacao || tipo == domain.TipoAcesso || tipo == domain.TipoPortabilidade || tipo == domain.TipoInformacaoCompartilhamento
}

func podeGerenciarPrivacidade(papel domain.Role) bool {
	return papel == domain.RoleGestor || papel == domain.RoleCorretorSolo || papel == domain.RoleSuperAdmin
}

func hashOpcional(valor string) *string {
	if valor == "" {
		return nil
	}
	soma := sha256.Sum256([]byte(valor))
	hash := hex.EncodeToString(soma[:])
	return &hash
}

func gerarProtocoloTitular(agora time.Time) (string, error) {
	aleatorio := make([]byte, 6)
	if _, err := rand.Read(aleatorio); err != nil {
		return "", fmt.Errorf("gerar protocolo seguro: %w", err)
	}
	return "KPT-" + agora.UTC().Format("20060102") + "-" + strings.ToUpper(hex.EncodeToString(aleatorio)), nil
}
