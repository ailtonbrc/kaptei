package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

var formatoNumeroWhatsApp = regexp.MustCompile(`^\+?[0-9]{8,20}$`)

type integracaoWhatsAppService struct {
	repositorio        ports.IntegracaoWhatsAppRepository
	protetor           ports.ProtetorSegredos
	disponivelServidor bool
	segredoAplicativo  string
	tokenVerificacao   string
	preparador         ports.PreparadorWhatsAppIntegracao
	preparadorOutbox   ports.PreparadorWhatsAppOutbox
	leads              ports.LeadRepository
}

func NewIntegracaoWhatsAppService(repositorio ports.IntegracaoWhatsAppRepository, protetor ports.ProtetorSegredos, segredoAplicativo, tokenVerificacao string, preparador ports.PreparadorWhatsAppIntegracao, preparadorOutbox ports.PreparadorWhatsAppOutbox, leads ports.LeadRepository) ports.IntegracaoWhatsAppService {
	disponivel := strings.TrimSpace(segredoAplicativo) != "" && strings.TrimSpace(tokenVerificacao) != ""
	return &integracaoWhatsAppService{repositorio: repositorio, protetor: protetor, disponivelServidor: disponivel,
		segredoAplicativo: strings.TrimSpace(segredoAplicativo), tokenVerificacao: strings.TrimSpace(tokenVerificacao),
		preparador: preparador, preparadorOutbox: preparadorOutbox, leads: leads}
}

func (s *integracaoWhatsAppService) ListarConversas(ctx context.Context, contaID, usuarioID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.ConversaWhatsApp], error) {
	filtro = normalizarFiltroPaginacao(filtro)
	var responsavelID *string
	if papel == domain.RoleCorretorEquipe {
		responsavelID = &usuarioID
	}
	return s.repositorio.ListarConversas(ctx, contaID, responsavelID, filtro)
}

func (s *integracaoWhatsAppService) ListarMensagens(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.MensagemWhatsApp], error) {
	if _, err := s.obterConversaAutorizada(ctx, conversaID, contaID, usuarioID, papel); err != nil {
		return nil, err
	}
	filtro = normalizarFiltroPaginacao(filtro)
	resultado, err := s.repositorio.ListarMensagens(ctx, conversaID, contaID, filtro)
	if err != nil {
		return nil, err
	}
	for _, mensagem := range resultado.Dados {
		if mensagem == nil {
			continue
		}
		if mensagem.Direcao == "ENTRADA" {
			recebida, erro := s.preparador.DecodificarConteudo(mensagem.ConteudoProtegido)
			if erro != nil {
				return nil, fmt.Errorf("revelar mensagem recebida: %w", erro)
			}
			mensagem.Conteudo = recebida.Texto
		} else {
			envio, erro := s.preparadorOutbox.DecodificarConteudo(mensagem.ConteudoProtegido)
			if erro != nil {
				return nil, fmt.Errorf("revelar mensagem enviada: %w", erro)
			}
			mensagem.Conteudo = envio.Texto
			if envio.Tipo == "TEMPLATE" {
				mensagem.Conteudo = "Template: " + envio.TemplateNome
			}
		}
		mensagem.ConteudoProtegido = ""
	}
	return resultado, nil
}

var formatoTemplateWhatsApp = regexp.MustCompile(`^[a-z0-9_]{1,512}$`)
var formatoIdiomaWhatsApp = regexp.MustCompile(`^[a-z]{2,3}(?:_[A-Z]{2})?$`)

func (s *integracaoWhatsAppService) EnviarTexto(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role, texto string) error {
	conversa, err := s.obterConversaAutorizada(ctx, conversaID, contaID, usuarioID, papel)
	if err != nil {
		return err
	}
	if conversa.JanelaAtendimentoAte == nil || !time.Now().UTC().Before(*conversa.JanelaAtendimentoAte) {
		return errors.New("a janela de atendimento expirou; use um template aprovado e autorizado")
	}
	texto = strings.TrimSpace(texto)
	if texto == "" || len([]rune(texto)) > 4096 {
		return errors.New("mensagem deve possuir entre 1 e 4096 caracteres")
	}
	return s.criarSaida(ctx, conversa, domain.SolicitacaoEnvioWhatsApp{Tipo: "TEXTO", Texto: texto})
}

func (s *integracaoWhatsAppService) EnviarTemplate(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role, template domain.EnvioTemplateWhatsApp) error {
	conversa, err := s.obterConversaAutorizada(ctx, conversaID, contaID, usuarioID, papel)
	if err != nil {
		return err
	}
	if !conversa.ConsentimentoMarketing {
		return errors.New("contato sem consentimento registrado para mensagem iniciada pela empresa")
	}
	template.Nome = strings.TrimSpace(template.Nome)
	template.Idioma = strings.TrimSpace(template.Idioma)
	if !formatoTemplateWhatsApp.MatchString(template.Nome) || !formatoIdiomaWhatsApp.MatchString(template.Idioma) {
		return errors.New("nome ou idioma do template inválido")
	}
	if len(template.Parametros) > 10 {
		return errors.New("template excede o limite de parâmetros")
	}
	for indice, parametro := range template.Parametros {
		template.Parametros[indice] = strings.TrimSpace(parametro)
		if template.Parametros[indice] == "" || len([]rune(template.Parametros[indice])) > 1024 {
			return errors.New("parâmetro do template inválido")
		}
	}
	return s.criarSaida(ctx, conversa, domain.SolicitacaoEnvioWhatsApp{Tipo: "TEMPLATE", TemplateNome: template.Nome, TemplateIdioma: template.Idioma, Parametros: template.Parametros})
}

func (s *integracaoWhatsAppService) RegistrarConsentimento(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role, consentiu bool, origem, evidencia string) error {
	if _, err := s.obterConversaAutorizada(ctx, conversaID, contaID, usuarioID, papel); err != nil {
		return err
	}
	origem = strings.ToUpper(strings.TrimSpace(origem))
	evidencia = strings.TrimSpace(evidencia)
	if consentiu && (origem == "" || evidencia == "") {
		return errors.New("origem e evidência são obrigatórias para registrar consentimento")
	}
	if len([]rune(origem)) > 80 || len([]rune(evidencia)) > 500 {
		return errors.New("dados de consentimento excedem os limites")
	}
	return s.repositorio.RegistrarConsentimento(ctx, conversaID, contaID, consentiu, origem, evidencia)
}

func (s *integracaoWhatsAppService) obterConversaAutorizada(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role) (*domain.ConversaWhatsApp, error) {
	conversa, err := s.repositorio.ObterConversa(ctx, conversaID, contaID)
	if err != nil || conversa == nil {
		return nil, errors.New("conversa WhatsApp não encontrada")
	}
	if papel == domain.RoleSuperAdmin || papel == domain.RoleGestor || papel == domain.RoleCorretorSolo {
		return conversa, nil
	}
	if conversa.LeadID == nil {
		return nil, errors.New("conversa ainda não possui lead atribuído")
	}
	lead, err := s.leads.GetByID(ctx, *conversa.LeadID, contaID)
	if err != nil || lead == nil || lead.UsuarioID == nil || *lead.UsuarioID != usuarioID {
		return nil, errors.New("sem permissão para acessar esta conversa")
	}
	return conversa, nil
}

func (s *integracaoWhatsAppService) criarSaida(ctx context.Context, conversa *domain.ConversaWhatsApp, parcial domain.SolicitacaoEnvioWhatsApp) error {
	configuracao, err := s.repositorio.ObterPorConta(ctx, conversa.ContaID)
	if err != nil || configuracao == nil || !configuracao.Ativa {
		return errors.New("integração WhatsApp ativa não encontrada")
	}
	id, err := gerarUUIDSeguro()
	if err != nil {
		return fmt.Errorf("gerar identificador da mensagem: %w", err)
	}
	parcial.IDMensagem = id
	parcial.ContaID = conversa.ContaID
	parcial.ConversaID = conversa.ID
	parcial.NumeroTelefoneID = configuracao.NumeroTelefoneID
	parcial.Destinatario = conversa.NumeroContato
	evento, protegido, err := s.preparadorOutbox.PrepararEnvio(parcial)
	if err != nil {
		return err
	}
	return s.repositorio.CriarMensagemSaida(ctx, &parcial, protegido, evento)
}

func gerarUUIDSeguro() (string, error) {
	bytesUUID := make([]byte, 16)
	if _, err := rand.Read(bytesUUID); err != nil {
		return "", err
	}
	bytesUUID[6] = (bytesUUID[6] & 0x0f) | 0x40
	bytesUUID[8] = (bytesUUID[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytesUUID[0:4], bytesUUID[4:6], bytesUUID[6:8], bytesUUID[8:10], bytesUUID[10:16]), nil
}

func (s *integracaoWhatsAppService) VerificarWebhook(modo, token, desafio string) (string, error) {
	if !s.disponivelServidor {
		return "", ErrMetaIndisponivel
	}
	if strings.TrimSpace(modo) != "subscribe" || desafio == "" || !compararSegredo(token, s.tokenVerificacao) {
		return "", errors.New("verificação do webhook WhatsApp recusada")
	}
	return desafio, nil
}

func (s *integracaoWhatsAppService) ReceberWebhook(ctx context.Context, assinatura string, corpo []byte) error {
	if !s.disponivelServidor || s.preparador == nil {
		return ErrMetaIndisponivel
	}
	if !validarAssinaturaMeta(corpo, assinatura, s.segredoAplicativo) {
		return ErrMetaAssinaturaInvalida
	}
	var notificacao struct {
		Objeto   string `json:"object"`
		Entradas []struct {
			Alteracoes []struct {
				Campo string `json:"field"`
				Valor struct {
					Produto   string `json:"messaging_product"`
					Metadados struct {
						NumeroTelefoneID string `json:"phone_number_id"`
					} `json:"metadata"`
					Contatos []struct {
						Numero string `json:"wa_id"`
						Perfil struct {
							Nome string `json:"name"`
						} `json:"profile"`
					} `json:"contacts"`
					Mensagens []struct {
						De        string `json:"from"`
						ID        string `json:"id"`
						Timestamp string `json:"timestamp"`
						Tipo      string `json:"type"`
						Texto     struct {
							Corpo string `json:"body"`
						} `json:"text"`
						Botao struct {
							Texto string `json:"text"`
						} `json:"button"`
						Interativo struct {
							Botao struct {
								Titulo string `json:"title"`
							} `json:"button_reply"`
							Lista struct {
								Titulo string `json:"title"`
							} `json:"list_reply"`
						} `json:"interactive"`
					} `json:"messages"`
					Status []struct {
						ID        string `json:"id"`
						Estado    string `json:"status"`
						Timestamp string `json:"timestamp"`
						Erros     []struct {
							Codigo   int    `json:"code"`
							Titulo   string `json:"title"`
							Mensagem string `json:"message"`
							Dados    struct {
								Detalhes string `json:"details"`
							} `json:"error_data"`
						} `json:"errors"`
					} `json:"statuses"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(corpo, &notificacao); err != nil || notificacao.Objeto != "whatsapp_business_account" {
		return errors.New("payload do webhook WhatsApp inválido")
	}
	eventos := make([]*domain.EventoIntegracao, 0)
	for _, entrada := range notificacao.Entradas {
		for _, alteracao := range entrada.Alteracoes {
			if alteracao.Campo != "messages" || alteracao.Valor.Produto != "whatsapp" {
				continue
			}
			numeroTelefoneID := strings.TrimSpace(alteracao.Valor.Metadados.NumeroTelefoneID)
			if !formatoIDMeta.MatchString(numeroTelefoneID) {
				return errors.New("payload WhatsApp sem Phone Number ID válido")
			}
			configuracao, err := s.repositorio.ObterPorNumeroTelefone(ctx, numeroTelefoneID)
			if err != nil {
				return err
			}
			if configuracao == nil {
				continue
			}
			nomes := make(map[string]string, len(alteracao.Valor.Contatos))
			for _, contato := range alteracao.Valor.Contatos {
				nomes[contato.Numero] = strings.TrimSpace(contato.Perfil.Nome)
			}
			for _, recebida := range alteracao.Valor.Mensagens {
				recebida.ID = strings.TrimSpace(recebida.ID)
				recebida.De = strings.TrimPrefix(strings.TrimSpace(recebida.De), "+")
				if recebida.ID == "" || len(recebida.ID) > 255 || !formatoNumeroWhatsApp.MatchString(recebida.De) {
					return errors.New("mensagem WhatsApp com identificadores inválidos")
				}
				instante := time.Now().UTC()
				if segundos, erro := strconv.ParseInt(recebida.Timestamp, 10, 64); erro == nil && segundos > 0 {
					instante = time.Unix(segundos, 0).UTC()
				}
				tipo := limitarRunas(recebida.Tipo, 32)
				if tipo == "" {
					tipo = "unknown"
				}
				mensagem := domain.MensagemWhatsAppRecebida{IdentificadorExterno: recebida.ID, NumeroTelefoneID: numeroTelefoneID,
					NumeroContato: recebida.De, NomeContato: limitarRunas(nomes[recebida.De], 120), Tipo: tipo,
					Texto: textoMensagemWhatsApp(recebida.Tipo, recebida.Texto.Corpo, recebida.Botao.Texto, recebida.Interativo.Botao.Titulo, recebida.Interativo.Lista.Titulo), OcorridaEm: instante}
				evento, err := s.preparador.PrepararMensagem(configuracao.ContaID, mensagem)
				if err != nil {
					return err
				}
				eventos = append(eventos, evento)
			}
			for _, recebido := range alteracao.Valor.Status {
				estado := mapearStatusWhatsApp(recebido.Estado)
				if estado == "" || strings.TrimSpace(recebido.ID) == "" {
					continue
				}
				instante := time.Now().UTC()
				if segundos, erro := strconv.ParseInt(recebido.Timestamp, 10, 64); erro == nil && segundos > 0 {
					instante = time.Unix(segundos, 0).UTC()
				}
				codigo, detalhe := "", ""
				if len(recebido.Erros) > 0 {
					codigo = strconv.Itoa(recebido.Erros[0].Codigo)
					detalhe = strings.TrimSpace(recebido.Erros[0].Titulo + " " + recebido.Erros[0].Mensagem + " " + recebido.Erros[0].Dados.Detalhes)
				}
				if err := s.repositorio.AtualizarStatusMensagem(ctx, recebido.ID, estado, codigo, detalhe, instante); err != nil {
					return err
				}
			}
		}
	}
	return s.repositorio.Enfileirar(ctx, eventos)
}

func mapearStatusWhatsApp(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "sent":
		return "ENVIADA"
	case "delivered":
		return "ENTREGUE"
	case "read":
		return "LIDA"
	case "failed":
		return "FALHOU"
	default:
		return ""
	}
}

func textoMensagemWhatsApp(tipo, texto, botao, respostaBotao, respostaLista string) string {
	for _, valor := range []string{texto, botao, respostaBotao, respostaLista} {
		if valor = strings.TrimSpace(valor); valor != "" {
			return limitarRunas(valor, 10000)
		}
	}
	if tipo == "" {
		tipo = "desconhecido"
	}
	return "[conteúdo " + limitarRunas(tipo, 32) + " recebido]"
}

func (s *integracaoWhatsAppService) ObterConfiguracao(ctx context.Context, contaID string, papel domain.Role) (*domain.ConfiguracaoWhatsApp, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissão para administrar a integração WhatsApp")
	}
	configuracao, err := s.repositorio.ObterPorConta(ctx, contaID)
	if err != nil {
		return nil, err
	}
	if configuracao == nil {
		return &domain.ConfiguracaoWhatsApp{ContaID: contaID, DisponivelNoServidor: s.disponivelServidor}, nil
	}
	configuracao.TokenAcessoProtegido = ""
	configuracao.DisponivelNoServidor = s.disponivelServidor
	return configuracao, nil
}

func (s *integracaoWhatsAppService) SalvarConfiguracao(ctx context.Context, contaID string, papel domain.Role, atualizacao domain.AtualizacaoWhatsApp) (*domain.ConfiguracaoWhatsApp, error) {
	if !podeGerenciarIntegracao(papel) {
		return nil, errors.New("sem permissão para administrar a integração WhatsApp")
	}
	if atualizacao.Ativa && !s.disponivelServidor {
		return nil, ErrMetaIndisponivel
	}
	wabaID := strings.TrimSpace(atualizacao.WABAID)
	numeroID := strings.TrimSpace(atualizacao.NumeroTelefoneID)
	if !formatoIDMeta.MatchString(wabaID) || !formatoIDMeta.MatchString(numeroID) {
		return nil, errors.New("identificadores WABA ou do número de telefone inválidos")
	}
	numeroExibicao := textoOpcional(atualizacao.NumeroExibicao)
	if numeroExibicao != nil {
		compacto := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(*numeroExibicao)
		if !formatoNumeroWhatsApp.MatchString(compacto) {
			return nil, errors.New("número de exibição do WhatsApp inválido")
		}
		numeroExibicao = &compacto
	}
	existente, err := s.repositorio.ObterPorConta(ctx, contaID)
	if err != nil {
		return nil, err
	}
	tokenProtegido := ""
	if existente != nil {
		tokenProtegido = existente.TokenAcessoProtegido
	}
	token := strings.TrimSpace(atualizacao.TokenAcesso)
	if token != "" {
		if len(token) < 20 || len(token) > 4096 || strings.ContainsAny(token, "\r\n\t ") {
			return nil, errors.New("token de acesso do WhatsApp inválido")
		}
		tokenProtegido, err = s.protetor.Proteger(token)
		if err != nil {
			return nil, fmt.Errorf("proteger token do WhatsApp: %w", err)
		}
	}
	if tokenProtegido == "" {
		return nil, errors.New("informe o token de acesso do WhatsApp")
	}
	configuracao := &domain.ConfiguracaoWhatsApp{
		ContaID: contaID, WABAID: wabaID, NumeroTelefoneID: numeroID,
		NumeroExibicao: numeroExibicao, TokenAcessoProtegido: tokenProtegido,
		TokenAcessoConfigurado: true, DisponivelNoServidor: s.disponivelServidor, Ativa: atualizacao.Ativa,
	}
	if err := s.repositorio.Salvar(ctx, configuracao); err != nil {
		return nil, err
	}
	configuracao.TokenAcessoProtegido = ""
	return configuracao, nil
}
