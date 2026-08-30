package ports

import (
	"context"
	"github.com/msdev/kaptei/internal/core/domain"
	"time"
)

// ContaRepository define as operações de banco de dados para ContaSaaS
type ContaRepository interface {
	Create(ctx context.Context, conta *domain.ContaSaaS) error
	GetByID(ctx context.Context, id string) (*domain.ContaSaaS, error)
	GetByLeadToken(ctx context.Context, token string) (*domain.ContaSaaS, error)
	GetByLeadTokenHash(ctx context.Context, tokenHash string) (*domain.ContaSaaS, error)
	Update(ctx context.Context, conta *domain.ContaSaaS) error
	AtualizarEstrategiaLeads(ctx context.Context, contaID, estrategia string) error
	RotacionarTokenLeads(ctx context.Context, contaID, tokenHash, prefixo string) error
}

// CadastroRepository garante atomicidade ao criar o tenant e seu primeiro usuário.
type CadastroRepository interface {
	CriarContaEUsuario(ctx context.Context, conta *domain.ContaSaaS, usuario *domain.Usuario) error
}

// UsuarioRepository define as operações de banco de dados para Usuario
type UsuarioRepository interface {
	Create(ctx context.Context, usuario *domain.Usuario) error
	GetByEmail(ctx context.Context, email string) (*domain.Usuario, error)
	GetByGoogleID(ctx context.Context, googleID string) (*domain.Usuario, error)
	GetByID(ctx context.Context, id string) (*domain.Usuario, error)
	ListByContaID(ctx context.Context, contaID string) ([]*domain.Usuario, error)
	Update(ctx context.Context, usuario *domain.Usuario) error
	AtualizarSenha(ctx context.Context, usuarioID, senhaHash string) error
	VincularGoogle(ctx context.Context, usuarioID, googleID string, avatar *string) error
	AtualizarStatusEquipe(ctx context.Context, usuarioID, contaID, status string, limiteCorretores *int) error
}

type ConviteEquipeRepository interface {
	Criar(ctx context.Context, convite *domain.ConviteEquipe, limiteCorretores *int, evento *domain.EventoOutbox) error
	ListarPendentes(ctx context.Context, contaID string) ([]*domain.ConviteEquipe, error)
	Cancelar(ctx context.Context, conviteID, contaID string) error
	Aceitar(ctx context.Context, tokenHash, nome, senhaHash string) (*domain.Usuario, error)
}

type SessaoRepository interface {
	Criar(ctx context.Context, sessao *domain.SessaoUsuario) error
	EstaAtiva(ctx context.Context, sessaoID, usuarioID string) (bool, error)
	Revogar(ctx context.Context, sessaoID, usuarioID string) error
}

// ImovelRepository define as operações de banco de dados para Imovel
type ImovelRepository interface {
	Create(ctx context.Context, imovel *domain.Imovel) error
	GetByID(ctx context.Context, id, contaID string) (*domain.Imovel, error)
	ListByContaID(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Imovel], error)
	Update(ctx context.Context, imovel *domain.Imovel) error
	Delete(ctx context.Context, id, contaID string, eventos []*domain.EventoOutbox) error
	AddFoto(ctx context.Context, foto *domain.ImovelFoto) error
	GetFoto(ctx context.Context, fotoID, imovelID string) (*domain.ImovelFoto, error)
	DeleteFoto(ctx context.Context, fotoID, imovelID string, eventos []*domain.EventoOutbox) error
}

type SitePublicoRepository interface {
	GetBySlug(ctx context.Context, slug string) (*domain.SitePublico, error)
	GetByContaID(ctx context.Context, contaID string) (*domain.SitePublico, error)
	Salvar(ctx context.Context, site *domain.SitePublico) error
	ListarImoveis(ctx context.Context, contaID string, filtros domain.FiltrosCatalogoPublico) ([]*domain.ImovelPublico, int, error)
	GetImovelBySlug(ctx context.Context, contaID, slug string) (*domain.ImovelPublico, error)
	ListarRotasSitemap(ctx context.Context) ([]domain.RotaSitemap, error)
}

type DominioSiteRepository interface {
	ObterPorConta(ctx context.Context, contaID string) (*domain.DominioSite, error)
	SalvarPendente(ctx context.Context, dominio *domain.DominioSite) error
	Ativar(ctx context.Context, id, contaID, tokenVerificacao string) error
	RegistrarFalha(ctx context.Context, id, contaID, tokenVerificacao, mensagem string) error
	ObterSitePorHostname(ctx context.Context, hostname string) (*domain.SitePublico, error)
}

// ConfiguracaoRepository define as operações para configuracoes_sistema
type ConfiguracaoRepository interface {
	Get(ctx context.Context, chave string) (*domain.ConfiguracaoSistema, error)
	Set(ctx context.Context, config *domain.ConfiguracaoSistema) error
}

// RecuperacaoSenhaRepository define as operações para gerenciar tokens
type RecuperacaoSenhaRepository interface {
	CreateToken(ctx context.Context, token *domain.RecuperacaoSenhaToken, evento *domain.EventoOutbox) error
	GetToken(ctx context.Context, token string) (*domain.RecuperacaoSenhaToken, error)
	ConsumirEAtualizarSenha(ctx context.Context, tokenID, usuarioID, senhaHash string) error
}

type OutboxRepository interface {
	Reservar(ctx context.Context, trabalhadorID string, limite int, duracaoBloqueio time.Duration) ([]*domain.EventoOutbox, error)
	Concluir(ctx context.Context, eventoID, trabalhadorID string) error
	Falhar(ctx context.Context, eventoID, trabalhadorID, mensagem string, proximaTentativa time.Time, definitivo bool) error
}

type IntegracaoMetaRepository interface {
	ObterPorConta(ctx context.Context, contaID string) (*domain.ConfiguracaoMetaLeads, error)
	ObterPorPagina(ctx context.Context, paginaID string) (*domain.ConfiguracaoMetaLeads, error)
	Salvar(ctx context.Context, configuracao *domain.ConfiguracaoMetaLeads) error
	Enfileirar(ctx context.Context, eventos []*domain.EventoIntegracao) error
	Reservar(ctx context.Context, trabalhadorID string, limite int, duracaoBloqueio time.Duration) ([]*domain.EventoIntegracao, error)
	Concluir(ctx context.Context, eventoID, trabalhadorID string) error
	Falhar(ctx context.Context, eventoID, trabalhadorID, mensagem string, proximaTentativa time.Time, definitivo bool) error
}

type IntegracaoWhatsAppRepository interface {
	ObterPorConta(ctx context.Context, contaID string) (*domain.ConfiguracaoWhatsApp, error)
	ObterPorNumeroTelefone(ctx context.Context, numeroTelefoneID string) (*domain.ConfiguracaoWhatsApp, error)
	Salvar(ctx context.Context, configuracao *domain.ConfiguracaoWhatsApp) error
	Enfileirar(ctx context.Context, eventos []*domain.EventoIntegracao) error
	Reservar(ctx context.Context, trabalhadorID string, limite int, duracaoBloqueio time.Duration) ([]*domain.EventoIntegracao, error)
	RegistrarMensagem(ctx context.Context, contaID string, mensagem *domain.MensagemWhatsAppRecebida, conteudoProtegido, chaveLead string) error
	Concluir(ctx context.Context, eventoID, trabalhadorID string) error
	Falhar(ctx context.Context, eventoID, trabalhadorID, mensagem string, proximaTentativa time.Time, definitivo bool) error
	ObterConversa(ctx context.Context, conversaID, contaID string) (*domain.ConversaWhatsApp, error)
	ListarConversas(ctx context.Context, contaID string, usuarioID *string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.ConversaWhatsApp], error)
	ListarMensagens(ctx context.Context, conversaID, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.MensagemWhatsApp], error)
	CriarMensagemSaida(ctx context.Context, solicitacao *domain.SolicitacaoEnvioWhatsApp, conteudoProtegido string, evento *domain.EventoOutbox) error
	MarcarMensagemEnviada(ctx context.Context, mensagemID, identificadorExterno string) error
	ObterIdentificadorMensagemSaida(ctx context.Context, mensagemID string) (*string, error)
	AtualizarStatusMensagem(ctx context.Context, identificadorExterno, status, codigoErro, detalheErro string, ocorridoEm time.Time) error
	RegistrarConsentimento(ctx context.Context, conversaID, contaID string, consentiu bool, origem, evidencia string) error
}

type PrivacidadeRepository interface {
	ObterContaPorSlug(ctx context.Context, slug string) (string, error)
	Criar(ctx context.Context, solicitacao *domain.SolicitacaoTitular) error
	Listar(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.SolicitacaoTitular], error)
	Obter(ctx context.Context, id, contaID string) (*domain.SolicitacaoTitular, error)
	VerificarIdentidade(ctx context.Context, id, contaID, usuarioID, metodo, evidenciaProtegida string) error
	Decidir(ctx context.Context, id, contaID, usuarioID, decisao, fundamento, observacao string) error
	GerarDadosExportacao(ctx context.Context, contaID, email, telefone string) (*domain.DadosTitularPersistidos, error)
	ConcluirExportacao(ctx context.Context, id, contaID, usuarioID string) error
	ExecutarDireito(ctx context.Context, id, contaID, usuarioID, email, telefone string, tipo domain.TipoSolicitacaoTitular) error
}

type RetencaoRepository interface {
	ObterPolitica(ctx context.Context, contaID string) (*domain.PoliticaRetencao, error)
	SalvarPolitica(ctx context.Context, politica *domain.PoliticaRetencao, usuarioID string) error
	GerarRelatorio(ctx context.Context, contaID string, politica *domain.PoliticaRetencao) (*domain.RelatorioRetencao, error)
	Executar(ctx context.Context, contaID, usuarioID string, politica *domain.PoliticaRetencao) (*domain.ResultadoRetencao, error)
	ListarBloqueios(ctx context.Context, contaID string) ([]domain.BloqueioRetencao, error)
	SalvarBloqueio(ctx context.Context, bloqueio *domain.BloqueioRetencao, usuarioID string) error
	RemoverBloqueio(ctx context.Context, id, contaID string) error
}

type PortalImobiliarioRepository interface {
	ObterConfiguracao(ctx context.Context, contaID, portal string) (*domain.ConfiguracaoPortal, error)
	SalvarConfiguracao(ctx context.Context, configuracao *domain.ConfiguracaoPortal, usuarioID string) error
	RotacionarToken(ctx context.Context, contaID, portal, hash, prefixo, usuarioID string) error
	ObterContaPorToken(ctx context.Context, portal, hash string) (string, error)
	ListarPublicacoes(ctx context.Context, contaID, portal string) ([]domain.PublicacaoPortal, error)
	SalvarPublicacao(ctx context.Context, contaID, portal, imovelID, usuarioID string, atualizacao domain.AtualizacaoPublicacaoPortal) error
	ObterDadosFeed(ctx context.Context, contaID, portal string) (*domain.DadosFeedPortal, error)
	ObterImovelDaConta(ctx context.Context, contaID, imovelID string) (*string, error)
}

// PlanoRepository define as operações de banco de dados para Plano
type PlanoRepository interface {
	ListarAtivos(ctx context.Context) ([]domain.Plano, error)
	GetByCodigo(ctx context.Context, codigo string) (*domain.Plano, error)
	AtualizarGatewayPriceID(ctx context.Context, codigo, priceID string) error
}

// ClienteRepository define as operações de banco de dados para Cliente
type ClienteRepository interface {
	Create(ctx context.Context, cliente *domain.Cliente) error
	GetByID(ctx context.Context, id, contaID string) (*domain.Cliente, error)
	ListByContaID(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Cliente], error)
	Update(ctx context.Context, cliente *domain.Cliente) error
	Delete(ctx context.Context, id, contaID string) error
	GetUltimoCorretorAtribuido(ctx context.Context, contaID string) (*string, error)
}

// LeadRepository define as operações de banco de dados para Lead
type LeadRepository interface {
	Create(ctx context.Context, lead *domain.Lead) error
	CreateDistribuido(ctx context.Context, lead *domain.Lead) error
	GetByID(ctx context.Context, id, contaID string) (*domain.Lead, error)
	ListByContaID(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Lead], error)
	Update(ctx context.Context, lead *domain.Lead) error
	Delete(ctx context.Context, id, contaID string) error
	Qualificar(ctx context.Context, id, contaID string) error
}

type MetricasConversaoRepository interface {
	Registrar(ctx context.Context, evento *domain.EventoConversao) error
	ObterResumo(ctx context.Context, contaID string, desde time.Time) (*domain.ResumoConversaoSite, error)
	ExpurgarExpirados(ctx context.Context) (int64, error)
}

// DashboardRepository define consultas analiticas
type DashboardRepository interface {
	GetFunilConversao(ctx context.Context, contaID string, usuarioID *string) (map[string]int, error)
	GetOrigemLeads(ctx context.Context, contaID string, usuarioID *string) (map[string]int, error)
	GetMetricasResumo(ctx context.Context, contaID string, usuarioID *string) (map[string]interface{}, error)
	GetEvolucaoLeads(ctx context.Context, contaID string, usuarioID *string) ([]string, []int, error)
}

// AgendamentoRepository define operações de banco de dados para Agendamentos
type AgendamentoRepository interface {
	Create(ctx context.Context, agendamento *domain.Agendamento) error
	GetByID(ctx context.Context, id, contaID string) (*domain.Agendamento, error)
	List(ctx context.Context, contaID string, usuarioID *string, inicio, fim time.Time) ([]*domain.Agendamento, error)
	Update(ctx context.Context, agendamento *domain.Agendamento) error
	Delete(ctx context.Context, id, contaID string) error
}
