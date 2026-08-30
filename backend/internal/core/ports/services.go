package ports

import (
	"context"
	"github.com/msdev/kaptei/internal/core/domain"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, email, senha string) (string, error)
	Register(ctx context.Context, nome, email, senha, tipoConta, plano string) (string, error)
	GoogleLogin(ctx context.Context, googleToken string, tipoConta string, plano string) (string, error)
	ValidateToken(ctx context.Context, token string) (*domain.Usuario, error)
	SolicitarRecuperacaoSenha(ctx context.Context, email string) error
	RedefinirSenha(ctx context.Context, token, novaSenha string) error
	Logout(ctx context.Context, token string) error
}

type EquipeService interface {
	Listar(ctx context.Context, contaID string, papelAtor domain.Role) ([]*domain.Usuario, []*domain.ConviteEquipe, error)
	Convidar(ctx context.Context, contaID, usuarioAtorID, email string, papelAtor domain.Role) error
	CancelarConvite(ctx context.Context, conviteID, contaID string, papelAtor domain.Role) error
	AtualizarStatus(ctx context.Context, usuarioID, contaID, usuarioAtorID, status string, papelAtor domain.Role) error
	AceitarConvite(ctx context.Context, token, nome, senha string) error
}

type UsuarioService interface {
	AtualizarPerfil(ctx context.Context, usuarioID string, dados domain.AtualizacaoPerfil) (*domain.Usuario, error)
	AlterarSenha(ctx context.Context, usuarioID, senhaAtual, novaSenha string) error
}

type ContaService interface {
	Obter(ctx context.Context, contaID string, papel domain.Role) (*domain.ContaSaaS, error)
	AtualizarEstrategiaLeads(ctx context.Context, contaID, estrategia string, papel domain.Role) error
	RotacionarTokenLeads(ctx context.Context, contaID string, papel domain.Role) (string, error)
}

// ImovelService define as operações de negócio para Imóveis
type ImovelService interface {
	Create(ctx context.Context, imovel *domain.Imovel) error
	GetByID(ctx context.Context, id, contaID string) (*domain.Imovel, error)
	List(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Imovel], error)
	Update(ctx context.Context, imovel *domain.Imovel) error
	Delete(ctx context.Context, id, contaID string) error
	AddFoto(ctx context.Context, imovelID, url string, isCapa bool) (*domain.ImovelFoto, error)
	UploadFoto(ctx context.Context, imovelID, contaID string, conteudo []byte, isCapa bool) (*domain.ImovelFoto, error)
	DeleteFoto(ctx context.Context, fotoID, imovelID, contaID string) error
}

type SitePublicoService interface {
	GetPublico(ctx context.Context, slug string) (*domain.SitePublico, error)
	GetAdministracao(ctx context.Context, contaID string) (*domain.SitePublico, error)
	Salvar(ctx context.Context, site *domain.SitePublico) error
	ListarImoveis(ctx context.Context, slug string, filtros domain.FiltrosCatalogoPublico) ([]*domain.ImovelPublico, int, error)
	GetImovel(ctx context.Context, slugSite, slugImovel string) (*domain.ImovelPublico, error)
	CapturarLead(ctx context.Context, slug string, captura domain.CapturaLeadPublico) error
	ListarRotasSitemap(ctx context.Context) ([]domain.RotaSitemap, error)
}

type DominioSiteService interface {
	Obter(ctx context.Context, contaID string, papel domain.Role) (*domain.DominioSite, error)
	Configurar(ctx context.Context, contaID string, papel domain.Role, hostname string) (*domain.DominioSite, error)
	Verificar(ctx context.Context, contaID string, papel domain.Role) (*domain.DominioSite, error)
	ResolverPublico(ctx context.Context, hostname string) (*domain.SitePublico, error)
}

type ResolvedorDNS interface {
	ConsultarTXT(ctx context.Context, nome string) ([]string, error)
}

// EmailService define a interface para envio de e-mails
type EmailService interface {
	SendEmail(ctx context.Context, idMensagem, to, subject, body string) error
}

// PreparadorEmailOutbox transforma dados de e-mail em um evento opaco e protegido.
// O caso de uso recebe o evento pronto, mas não conhece criptografia nem persistência.
type PreparadorEmailOutbox interface {
	PrepararEmail(contaID *string, chaveIdempotencia, destinatario, assunto, corpoHTML string) (*domain.EventoOutbox, error)
	DecodificarEmail(evento *domain.EventoOutbox) (*domain.MensagemEmail, error)
}

type PreparadorObjetoOutbox interface {
	PrepararExclusaoObjeto(contaID, chaveIdempotencia, provedor, chave string) (*domain.EventoOutbox, error)
	DecodificarExclusaoObjeto(evento *domain.EventoOutbox) (*domain.SolicitacaoExclusaoObjeto, error)
}

type ProcessadorImagem interface {
	Processar(conteudo []byte) (*domain.ImagemProcessada, error)
}

type ArmazenamentoObjetos interface {
	Salvar(ctx context.Context, chave string, conteudo []byte, tipoConteudo string) (string, error)
	Excluir(ctx context.Context, chave string) error
	Nome() string
}

// ConfiguracaoService define as operações para configurações
type ConfiguracaoService interface {
	GetConfig(ctx context.Context, chave string) (*domain.ConfiguracaoSistema, error)
	UpdateConfig(ctx context.Context, chave string, valor interface{}, descricao string) error
}

// ProtetorSegredos abstrai a criptografia de valores sensíveis persistidos.
type ProtetorSegredos interface {
	Proteger(valor string) (string, error)
	Revelar(valor string) (string, error)
}

type ClienteMetaGraph interface {
	ObterLead(ctx context.Context, leadID, tokenPagina string) (*domain.LeadMeta, error)
}

type ErroNaoRetentavel interface {
	error
	NaoRetentavel() bool
}

type IntegracaoMetaService interface {
	ObterConfiguracao(ctx context.Context, contaID string, papel domain.Role) (*domain.ConfiguracaoMetaLeads, error)
	SalvarConfiguracao(ctx context.Context, contaID string, papel domain.Role, atualizacao domain.AtualizacaoMetaLeads) (*domain.ConfiguracaoMetaLeads, error)
	VerificarWebhook(modo, token, desafio string) (string, error)
	ReceberWebhook(ctx context.Context, assinatura string, corpo []byte) error
}

type IntegracaoWhatsAppService interface {
	ObterConfiguracao(ctx context.Context, contaID string, papel domain.Role) (*domain.ConfiguracaoWhatsApp, error)
	SalvarConfiguracao(ctx context.Context, contaID string, papel domain.Role, atualizacao domain.AtualizacaoWhatsApp) (*domain.ConfiguracaoWhatsApp, error)
	VerificarWebhook(modo, token, desafio string) (string, error)
	ReceberWebhook(ctx context.Context, assinatura string, corpo []byte) error
	ListarConversas(ctx context.Context, contaID, usuarioID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.ConversaWhatsApp], error)
	ListarMensagens(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.MensagemWhatsApp], error)
	EnviarTexto(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role, texto string) error
	EnviarTemplate(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role, template domain.EnvioTemplateWhatsApp) error
	RegistrarConsentimento(ctx context.Context, conversaID, contaID, usuarioID string, papel domain.Role, consentiu bool, origem, evidencia string) error
}

type PrivacidadeService interface {
	CriarPublica(ctx context.Context, slug string, nova domain.NovaSolicitacaoTitular) (string, error)
	Listar(ctx context.Context, contaID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.SolicitacaoTitular], error)
	Obter(ctx context.Context, id, contaID string, papel domain.Role) (*domain.SolicitacaoTitular, error)
	VerificarIdentidade(ctx context.Context, id, contaID, usuarioID string, papel domain.Role, metodo, evidencia string) error
	Decidir(ctx context.Context, id, contaID, usuarioID string, papel domain.Role, decisao domain.DecisaoSolicitacaoTitular) error
	Exportar(ctx context.Context, id, contaID, usuarioID string, papel domain.Role) (*domain.ExportacaoTitular, error)
	Executar(ctx context.Context, id, contaID, usuarioID string, papel domain.Role) error
}

type RetencaoService interface {
	ObterPolitica(ctx context.Context, contaID string, papel domain.Role) (*domain.PoliticaRetencao, error)
	SalvarPolitica(ctx context.Context, contaID, usuarioID string, papel domain.Role, politica domain.PoliticaRetencao) error
	GerarRelatorio(ctx context.Context, contaID string, papel domain.Role) (*domain.RelatorioRetencao, error)
	Executar(ctx context.Context, contaID, usuarioID string, papel domain.Role, confirmacao string) (*domain.ResultadoRetencao, error)
	ListarBloqueios(ctx context.Context, contaID string, papel domain.Role) ([]domain.BloqueioRetencao, error)
	SalvarBloqueio(ctx context.Context, contaID, usuarioID string, papel domain.Role, bloqueio domain.BloqueioRetencao) (*domain.BloqueioRetencao, error)
	RemoverBloqueio(ctx context.Context, id, contaID string, papel domain.Role) error
}

type GeradorFeedPortal interface {
	Validar(dados *domain.DadosFeedPortal) *domain.DiagnosticoFeedPortal
	Gerar(dados *domain.DadosFeedPortal, origemPublica string, instante time.Time) ([]byte, error)
}

type PortalImobiliarioService interface {
	ObterConfiguracao(ctx context.Context, contaID string, papel domain.Role) (*domain.ConfiguracaoPortal, error)
	SalvarConfiguracao(ctx context.Context, contaID, usuarioID string, papel domain.Role, configuracao domain.ConfiguracaoPortal) (*domain.ConfiguracaoPortal, error)
	RotacionarToken(ctx context.Context, contaID, usuarioID string, papel domain.Role) (*domain.CredencialFeedPortal, error)
	ListarPublicacoes(ctx context.Context, contaID string, papel domain.Role) ([]domain.PublicacaoPortal, error)
	SalvarPublicacao(ctx context.Context, contaID, imovelID, usuarioID string, papel domain.Role, atualizacao domain.AtualizacaoPublicacaoPortal) error
	Diagnosticar(ctx context.Context, contaID string, papel domain.Role) (*domain.DiagnosticoFeedPortal, error)
	GerarFeedPublico(ctx context.Context, token string) ([]byte, error)
	ReceberLead(ctx context.Context, token, autorizacao string, lead domain.LeadGrupoOLX) error
}

type PreparadorWhatsAppIntegracao interface {
	PrepararMensagem(contaID string, mensagem domain.MensagemWhatsAppRecebida) (*domain.EventoIntegracao, error)
	DecodificarMensagem(evento *domain.EventoIntegracao) (*domain.MensagemWhatsAppRecebida, error)
	DecodificarConteudo(conteudoProtegido string) (*domain.MensagemWhatsAppRecebida, error)
}

type PreparadorWhatsAppOutbox interface {
	PrepararEnvio(solicitacao domain.SolicitacaoEnvioWhatsApp) (*domain.EventoOutbox, string, error)
	DecodificarEnvio(evento *domain.EventoOutbox) (*domain.SolicitacaoEnvioWhatsApp, error)
	DecodificarConteudo(conteudoProtegido string) (*domain.SolicitacaoEnvioWhatsApp, error)
}

type ClienteWhatsApp interface {
	Enviar(ctx context.Context, numeroTelefoneID, tokenAcesso string, solicitacao *domain.SolicitacaoEnvioWhatsApp) (string, error)
}

type TratadorEventoOutbox interface {
	Tipo() string
	Processar(ctx context.Context, evento *domain.EventoOutbox) error
}

// ClienteService define as operações de negócio para Clientes (CRM)
type ClienteService interface {
	Create(ctx context.Context, cliente *domain.Cliente, usuarioAtorID string, papel domain.Role) error
	GetByID(ctx context.Context, id, contaID, usuarioAtorID string, papel domain.Role) (*domain.Cliente, error)
	List(ctx context.Context, contaID, usuarioAtorID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Cliente], error)
	Update(ctx context.Context, cliente *domain.Cliente, usuarioAtorID string, papel domain.Role) error
	Delete(ctx context.Context, id, contaID, usuarioAtorID string, papel domain.Role) error
}

// LeadService define as operações de negócio para Leads
type LeadService interface {
	ProcessarWebhook(ctx context.Context, token string, captura domain.CapturaLeadWebhook) error
	CaptarSite(ctx context.Context, contaID string, captura domain.CapturaLeadPublico, imovelID *string) error
	CaptarIntegracao(ctx context.Context, contaID string, captura domain.CapturaLeadIntegracao) error
	List(ctx context.Context, contaID, usuarioID string, papel domain.Role, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Lead], error)
	Atribuir(ctx context.Context, leadID, contaID, usuarioAtorID, usuarioDestinoID string, papel domain.Role) error
	Qualificar(ctx context.Context, leadID, contaID, usuarioAtorID string, papel domain.Role) error
	Descartar(ctx context.Context, leadID, contaID, usuarioAtorID, motivo string, papel domain.Role) error
}

type MetricasConversaoService interface {
	Registrar(ctx context.Context, slug string, evento domain.EventoConversaoPublico) error
}

// DashboardService define as operacoes de negocio para Dashboard
type DashboardService interface {
	GetDashboardPremium(ctx context.Context, contaID, usuarioID string, papel domain.Role) (map[string]interface{}, error)
}

// AgendamentoService define as operacoes de negocio para Agendamentos
type AgendamentoService interface {
	Create(ctx context.Context, agendamento *domain.Agendamento, usuarioAtorID string, papel domain.Role) error
	GetByID(ctx context.Context, id, contaID string) (*domain.Agendamento, error)
	List(ctx context.Context, contaID string, usuarioID *string, inicio, fim time.Time) ([]*domain.Agendamento, error)
	Update(ctx context.Context, agendamento *domain.Agendamento) error
	Delete(ctx context.Context, id, contaID string) error
}
