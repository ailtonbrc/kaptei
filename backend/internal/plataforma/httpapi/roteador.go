package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/adapters/gateways"
	"github.com/msdev/kaptei/internal/adapters/handlers"
	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/adapters/portais/vrsync"
	"github.com/msdev/kaptei/internal/adapters/repositories"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"github.com/msdev/kaptei/internal/core/services"
	"github.com/msdev/kaptei/internal/plataforma/configuracao"
	"github.com/rs/cors"
)

type DependenciasRoteador struct {
	ProtetorSegredos         ports.ProtetorSegredos
	PreparadorEmail          ports.PreparadorEmailOutbox
	ProcessadorImagem        ports.ProcessadorImagem
	Objetos                  ports.ArmazenamentoObjetos
	PreparadorObjeto         ports.PreparadorObjetoOutbox
	PreparadorWhatsApp       ports.PreparadorWhatsAppIntegracao
	PreparadorWhatsAppOutbox ports.PreparadorWhatsAppOutbox
	Metricas                 ports.MetricasAplicacao
	HandlerMetricas          http.Handler
	PrefixoMidiaLocal        string
	HandlerMidiaLocal        http.Handler
}

func NovoRoteador(banco *sql.DB, ambiente configuracao.Ambiente, dependencias DependenciasRoteador) http.Handler {
	contaRepo := repositories.NewContaRepository(banco)
	cadastroRepo := repositories.NewCadastroRepository(banco)
	usuarioRepo := repositories.NewUsuarioRepository(banco)
	planoRepo := repositories.NewPlanoRepository(banco)
	configRepo := repositories.NewConfiguracaoRepository(banco)
	recuperacaoRepo := repositories.NewRecuperacaoSenhaRepository(banco)
	dashboardRepo := repositories.NewDashboardPostgres(banco)
	imovelRepo := repositories.NewImovelPostgres(banco)
	clienteRepo := repositories.NewClientePostgres(banco)
	interacaoRepo := repositories.NewInteracaoPostgres(banco)
	leadRepo := repositories.NewLeadPostgres(banco)
	agendamentoRepo := repositories.NewAgendamentoPostgres(banco)
	sitePublicoRepo := repositories.NewSitePublicoPostgres(banco)
	billingRepo := repositories.NewBillingPostgres(banco)
	conviteEquipeRepo := repositories.NewConviteEquipeRepository(banco)
	sessaoRepo := repositories.NewSessaoRepository(banco)
	metaRepo := repositories.NewIntegracaoMetaPostgres(banco)
	whatsAppRepo := repositories.NewIntegracaoWhatsAppPostgres(banco)
	privacidadeRepo := repositories.NewPrivacidadePostgres(banco)
	metricasConversaoRepo := repositories.NewMetricasConversaoPostgres(banco)
	dominioRepo := repositories.NewDominioSitePostgres(banco)
	retencaoRepo := repositories.NewRetencaoPostgres(banco)

	portalRepo := repositories.NewPortalImobiliarioPostgres(banco)
	authSvc := services.NewAuthService(usuarioRepo, contaRepo, cadastroRepo, recuperacaoRepo, sessaoRepo, dependencias.PreparadorEmail, configRepo, planoRepo, ambiente.JWTSecret, ambiente.URLPublicaAplicacao)
	dashboardSvc := services.NewDashboardService(dashboardRepo, metricasConversaoRepo)
	configSvc := services.NewConfiguracaoService(configRepo, dependencias.ProtetorSegredos)
	imovelSvc := services.NewImovelService(imovelRepo, dependencias.ProcessadorImagem, dependencias.Objetos, dependencias.PreparadorObjeto)
	clienteSvc := services.NewClienteService(clienteRepo, usuarioRepo)
	interacaoSvc := services.NewInteracaoService(interacaoRepo, clienteRepo)
	leadSvc := services.NewLeadService(leadRepo, contaRepo, usuarioRepo)
	sitePublicoSvc := services.NewSitePublicoService(sitePublicoRepo, leadSvc)
	agendamentoSvc := services.NewAgendamentoService(agendamentoRepo, clienteRepo, usuarioRepo, imovelRepo)
	pagamentoGateway := gateways.NewStripeGateway(ambiente.StripeSecretKey, ambiente.StripeWebhookSecret)
	billingSvc := services.NewBillingService(contaRepo, usuarioRepo, planoRepo, pagamentoGateway, billingRepo, ambiente.URLPublicaAplicacao)
	equipeSvc := services.NewEquipeService(usuarioRepo, conviteEquipeRepo, contaRepo, planoRepo, dependencias.PreparadorEmail, ambiente.URLPublicaAplicacao)
	usuarioSvc := services.NewUsuarioService(usuarioRepo)
	contaSvc := services.NewContaService(contaRepo)
	metaSvc := services.NewIntegracaoMetaService(metaRepo, dependencias.ProtetorSegredos, ambiente.MetaAppSecret, ambiente.MetaWebhookVerifyToken, ambiente.OutboxMaxTentativas)
	whatsAppSvc := services.NewIntegracaoWhatsAppService(whatsAppRepo, dependencias.ProtetorSegredos, ambiente.MetaAppSecret, ambiente.MetaWebhookVerifyToken, dependencias.PreparadorWhatsApp, dependencias.PreparadorWhatsAppOutbox, leadRepo)
	privacidadeSvc := services.NewPrivacidadeService(privacidadeRepo, dependencias.ProtetorSegredos)
	dominioSvc := services.NewDominioSiteService(dominioRepo, gateways.NewResolvedorDNS(), ambiente.URLPublicaAplicacao)
	retencaoSvc := services.NewRetencaoService(retencaoRepo)
	metricasConversaoSvc := services.NewMetricasConversaoService(sitePublicoRepo, metricasConversaoRepo)

	portalSvc := services.NewPortalImobiliarioService(portalRepo, vrsync.NovoGerador(), leadSvc, ambiente.URLPublicaAplicacao, ambiente.GrupoOLXWebhookSecret)
	authHandler := handlers.NewAuthHandler(authSvc, contaRepo, ambiente.CookieSeguro)
	planoHandler := handlers.NewPlanoHandler(planoRepo)
	usuarioHandler := handlers.NewUsuarioHandler(usuarioSvc)
	configHandler := handlers.NewConfiguracaoHandler(configSvc)
	imovelHandler := handlers.NewImovelHandler(imovelSvc, int64(ambiente.ImagemMaximoBytes), ambiente.ImagemMaximoConcorrente)
	clienteHandler := handlers.NewClienteHandler(clienteSvc)
	interacaoHandler := handlers.NewInteracaoHandler(interacaoSvc)
	leadHandler := handlers.NewLeadHandler(leadSvc)
	agendamentoHandler := handlers.NewAgendamentoHandler(agendamentoSvc)
	dashboardHandler := handlers.NewDashboardHandler(dashboardSvc)
	contaHandler := handlers.NewContaHandler(contaSvc)
	billingHandler := handlers.NewBillingHandler(billingSvc)
	sitePublicoHandler := handlers.NewSitePublicoHandler(sitePublicoSvc, ambiente.URLPublicaAplicacao)
	equipeHandler := handlers.NewEquipeHandler(equipeSvc)
	metaHandler := handlers.NewIntegracaoMetaHandler(metaSvc)
	metricasConversaoHandler := handlers.NewMetricasConversaoHandler(metricasConversaoSvc)
	whatsAppHandler := handlers.NewIntegracaoWhatsAppHandler(whatsAppSvc)
	privacidadeHandler := handlers.NewPrivacidadeHandler(privacidadeSvc)
	dominioHandler := handlers.NewDominioSiteHandler(dominioSvc)
	retencaoHandler := handlers.NewRetencaoHandler(retencaoSvc)

	portalHandler := handlers.NewPortalImobiliarioHandler(portalSvc)
	publico := http.NewServeMux()
	publico.HandleFunc("GET /health", responderSaude)
	publico.HandleFunc("GET /ready", responderProntidao(banco))
	if dependencias.HandlerMetricas != nil {
		publico.Handle("GET /metrics", dependencias.HandlerMetricas)
	}
	publico.HandleFunc("GET /api/v1/planos", planoHandler.Listar)
	publico.HandleFunc("GET /sitemap.xml", sitePublicoHandler.Sitemap)
	publico.HandleFunc("GET /robots.txt", sitePublicoHandler.Robots)
	if dependencias.HandlerMidiaLocal != nil && dependencias.PrefixoMidiaLocal != "" {
		padrao := "GET " + strings.TrimRight(dependencias.PrefixoMidiaLocal, "/") + "/{arquivo...}"
		publico.Handle(padrao, dependencias.HandlerMidiaLocal)
	}
	limitarAutenticacao := middlewares.LimitarRequisicoes(ambiente.LimiteAutenticacao, ambiente.JanelaAutenticacao, ambiente.ConfiarProxy)
	limitarLeituraPublica := middlewares.LimitarRequisicoes(ambiente.LimiteLeituraPublica, ambiente.JanelaLeituraPublica, ambiente.ConfiarProxy)
	publico.Handle("POST /api/auth/login", limitarAutenticacao(http.HandlerFunc(authHandler.Login)))
	publico.Handle("POST /api/auth/google", limitarAutenticacao(http.HandlerFunc(authHandler.GoogleLogin)))
	publico.Handle("POST /api/auth/register", limitarAutenticacao(http.HandlerFunc(authHandler.Register)))
	publico.Handle("POST /api/auth/esqueci-senha", limitarAutenticacao(http.HandlerFunc(authHandler.EsqueciSenha)))
	publico.Handle("POST /api/auth/redefinir-senha", limitarAutenticacao(http.HandlerFunc(authHandler.RedefinirSenha)))
	publico.Handle("POST /api/auth/aceitar-convite", limitarAutenticacao(http.HandlerFunc(equipeHandler.AceitarConvite)))
	publico.Handle("POST /api/auth/logout", middlewares.ValidarOrigemCookie(ambiente.OrigensCORS)(http.HandlerFunc(authHandler.Logout)))
	publico.HandleFunc("POST /api/webhooks/stripe", billingHandler.Webhook)
	publico.HandleFunc("GET /api/webhooks/meta/leads", metaHandler.VerificarWebhook)
	publico.HandleFunc("POST /api/webhooks/meta/leads", metaHandler.ReceberWebhook)
	publico.HandleFunc("GET /api/webhooks/meta/whatsapp", whatsAppHandler.VerificarWebhook)
	publico.HandleFunc("POST /api/webhooks/meta/whatsapp", whatsAppHandler.ReceberWebhook)
	publico.HandleFunc("GET /api/public/configuracoes/{chave}", configHandler.GetPublicConfig)
	publico.Handle("POST /api/v1/webhooks/leads/{token}", middlewares.LimitarRequisicoes(
		ambiente.LimiteLeadsPublicos, ambiente.JanelaLeadsPublicos, ambiente.ConfiarProxy,
	)(http.HandlerFunc(leadHandler.ProcessarWebhook)))
	publico.Handle("GET /api/public/dominios/{hostname}", limitarLeituraPublica(http.HandlerFunc(dominioHandler.ResolverPublico)))
	publico.Handle("GET /api/public/portais/grupo-olx/{token}/vrsync.xml", limitarLeituraPublica(
		http.HandlerFunc(portalHandler.FeedVRSync),
	))
	publico.Handle("POST /api/webhooks/leads", middlewares.LimitarRequisicoes(
		ambiente.LimiteLeadsPublicos, ambiente.JanelaLeadsPublicos, ambiente.ConfiarProxy,
	)(http.HandlerFunc(leadHandler.ProcessarWebhookSeguro)))
	publico.Handle("POST /api/webhooks/portais/grupo-olx/{token}/leads", middlewares.LimitarRequisicoes(
		ambiente.LimiteLeadsPublicos, ambiente.JanelaLeadsPublicos, ambiente.ConfiarProxy,
	)(http.HandlerFunc(portalHandler.ReceberLead)))
	publico.Handle("GET /api/public/sites/{slug}", limitarLeituraPublica(http.HandlerFunc(sitePublicoHandler.GetPublico)))
	publico.Handle("POST /api/public/sites/{slug}/eventos-conversao", limitarLeituraPublica(
		http.HandlerFunc(metricasConversaoHandler.Registrar),
	))
	publico.Handle("GET /api/public/sites/{slug}/imoveis", limitarLeituraPublica(http.HandlerFunc(sitePublicoHandler.ListarImoveis)))
	publico.Handle("GET /api/public/sites/{slug}/imoveis/{slug_imovel}", limitarLeituraPublica(http.HandlerFunc(sitePublicoHandler.GetImovel)))
	publico.Handle("POST /api/public/sites/{slug}/leads", middlewares.LimitarRequisicoes(
		ambiente.LimiteLeadsPublicos, ambiente.JanelaLeadsPublicos, ambiente.ConfiarProxy,
	)(http.HandlerFunc(sitePublicoHandler.CapturarLead)))
	publico.Handle("POST /api/public/sites/{slug}/privacidade/solicitacoes", middlewares.LimitarRequisicoes(
		ambiente.LimiteLeadsPublicos, ambiente.JanelaLeadsPublicos, ambiente.ConfiarProxy,
	)(http.HandlerFunc(privacidadeHandler.CriarPublica)))

	protegido := http.NewServeMux()
	protegido.HandleFunc("GET /api/v1/sessao", authHandler.Sessao)
	protegido.HandleFunc("GET /api/v1/me", responderUsuarioAtual)
	protegido.HandleFunc("GET /api/v1/conta", contaHandler.GetMinhaConta)
	protegido.HandleFunc("PUT /api/v1/conta/leads", contaHandler.UpdateLeadConfig)
	protegido.HandleFunc("POST /api/v1/conta/leads/token", contaHandler.RotacionarTokenLeads)
	protegido.HandleFunc("GET /api/v1/integracoes/meta/leads", metaHandler.ObterConfiguracao)
	protegido.HandleFunc("PUT /api/v1/integracoes/meta/leads", metaHandler.SalvarConfiguracao)
	protegido.HandleFunc("GET /api/v1/integracoes/whatsapp", whatsAppHandler.ObterConfiguracao)
	protegido.HandleFunc("PUT /api/v1/integracoes/whatsapp", whatsAppHandler.SalvarConfiguracao)
	protegido.HandleFunc("GET /api/v1/integracoes/whatsapp/conversas", whatsAppHandler.ListarConversas)
	protegido.HandleFunc("GET /api/v1/integracoes/whatsapp/conversas/{id}/mensagens", whatsAppHandler.ListarMensagens)
	protegido.HandleFunc("POST /api/v1/integracoes/whatsapp/conversas/{id}/mensagens", whatsAppHandler.EnviarTexto)
	protegido.HandleFunc("POST /api/v1/integracoes/whatsapp/conversas/{id}/templates", whatsAppHandler.EnviarTemplate)
	protegido.HandleFunc("PUT /api/v1/integracoes/whatsapp/conversas/{id}/consentimento", whatsAppHandler.RegistrarConsentimento)
	protegido.HandleFunc("PUT /api/v1/usuarios/perfil", usuarioHandler.UpdatePerfil)
	protegido.HandleFunc("PUT /api/v1/usuarios/senha", usuarioHandler.UpdateSenha)
	protegido.HandleFunc("GET /api/v1/configuracoes/{chave}", configHandler.GetConfig)
	protegido.HandleFunc("PUT /api/v1/configuracoes/{chave}", configHandler.UpdateConfig)
	protegido.HandleFunc("POST /api/v1/billing/checkout", billingHandler.Checkout)
	protegido.HandleFunc("POST /api/v1/billing/portal", billingHandler.Portal)
	protegido.HandleFunc("GET /api/v1/planos/administracao", planoHandler.ListarAdministracao)
	protegido.HandleFunc("PUT /api/v1/planos/{codigo}/gateway", planoHandler.ConfigurarGateway)
	protegido.HandleFunc("GET /api/v1/site", sitePublicoHandler.GetAdministracao)
	protegido.HandleFunc("PUT /api/v1/site", sitePublicoHandler.Salvar)
	protegido.HandleFunc("GET /api/v1/site/dominio", dominioHandler.Obter)
	protegido.HandleFunc("PUT /api/v1/site/dominio", dominioHandler.Configurar)
	protegido.HandleFunc("POST /api/v1/site/dominio/verificacao", dominioHandler.Verificar)
	protegido.HandleFunc("GET /api/v1/equipe", equipeHandler.Listar)
	protegido.HandleFunc("POST /api/v1/equipe/convites", equipeHandler.Convidar)
	protegido.HandleFunc("DELETE /api/v1/equipe/convites/{id}", equipeHandler.CancelarConvite)
	protegido.HandleFunc("PATCH /api/v1/equipe/{id}/status", equipeHandler.AtualizarStatus)
	protegido.HandleFunc("GET /api/v1/privacidade/solicitacoes", privacidadeHandler.Listar)
	protegido.HandleFunc("GET /api/v1/privacidade/solicitacoes/{id}", privacidadeHandler.Obter)
	protegido.HandleFunc("POST /api/v1/privacidade/solicitacoes/{id}/verificacao", privacidadeHandler.VerificarIdentidade)
	protegido.HandleFunc("POST /api/v1/privacidade/solicitacoes/{id}/decisao", privacidadeHandler.Decidir)
	protegido.HandleFunc("POST /api/v1/privacidade/solicitacoes/{id}/exportacao", privacidadeHandler.Exportar)
	protegido.HandleFunc("POST /api/v1/privacidade/solicitacoes/{id}/execucao", privacidadeHandler.Executar)
	protegido.HandleFunc("GET /api/v1/privacidade/retencao/politica", retencaoHandler.ObterPolitica)
	protegido.HandleFunc("PUT /api/v1/privacidade/retencao/politica", retencaoHandler.SalvarPolitica)
	protegido.HandleFunc("GET /api/v1/privacidade/retencao/relatorio", retencaoHandler.Relatorio)
	protegido.HandleFunc("POST /api/v1/privacidade/retencao/execucao", retencaoHandler.Executar)
	protegido.HandleFunc("GET /api/v1/privacidade/retencao/bloqueios", retencaoHandler.ListarBloqueios)
	protegido.HandleFunc("POST /api/v1/privacidade/retencao/bloqueios", retencaoHandler.SalvarBloqueio)

	negocios := http.NewServeMux()
	protegido.HandleFunc("DELETE /api/v1/privacidade/retencao/bloqueios/{id}", retencaoHandler.RemoverBloqueio)
	protegido.HandleFunc("GET /api/v1/integracoes/portais/grupo-olx", portalHandler.ObterConfiguracao)
	protegido.HandleFunc("PUT /api/v1/integracoes/portais/grupo-olx", portalHandler.SalvarConfiguracao)
	protegido.HandleFunc("POST /api/v1/integracoes/portais/grupo-olx/token", portalHandler.RotacionarToken)
	protegido.HandleFunc("GET /api/v1/integracoes/portais/grupo-olx/publicacoes", portalHandler.ListarPublicacoes)
	protegido.HandleFunc("PUT /api/v1/integracoes/portais/grupo-olx/publicacoes/{id}", portalHandler.SalvarPublicacao)
	protegido.HandleFunc("GET /api/v1/integracoes/portais/grupo-olx/diagnostico", portalHandler.Diagnosticar)
	registrarRotasNegocios(negocios, dashboardHandler, imovelHandler, clienteHandler, agendamentoHandler, interacaoHandler, leadHandler)
	protegido.Handle("/api/v1/negocios/", middlewares.RequireActivePlan(contaRepo)(negocios))
	publico.Handle("/api/v1/", middlewares.ValidarOrigemCookie(ambiente.OrigensCORS)(
		middlewares.RequireAuth(authSvc)(middlewares.AuditarMutacoes(banco)(protegido)),
	))

	politicaCORS := cors.New(cors.Options{
		AllowedOrigins:   ambiente.OrigensCORS,
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Idempotency-Key", "X-Request-ID", "X-Kaptei-Token"},
		MaxAge:           600,
	})
	return middlewares.Observabilidade(politicaCORS.Handler(publico), dependencias.Metricas)
}

func registrarRotasNegocios(
	mux *http.ServeMux,
	dashboard *handlers.DashboardHandler,
	imovel *handlers.ImovelHandler,
	cliente *handlers.ClienteHandler,
	agendamento *handlers.AgendamentoHandler,
	interacao *handlers.InteracaoHandler,
	lead *handlers.LeadHandler,
) {
	mux.HandleFunc("GET /api/v1/negocios/dashboard/premium", dashboard.GetResumoPremium)

	mux.HandleFunc("GET /api/v1/negocios/imoveis", imovel.List)
	mux.HandleFunc("POST /api/v1/negocios/imoveis", imovel.Create)
	mux.HandleFunc("GET /api/v1/negocios/imoveis/{id}", imovel.GetByID)
	mux.HandleFunc("PUT /api/v1/negocios/imoveis/{id}", imovel.Update)
	mux.HandleFunc("DELETE /api/v1/negocios/imoveis/{id}", imovel.Delete)
	mux.HandleFunc("POST /api/v1/negocios/imoveis/{id}/fotos", imovel.AddFoto)
	mux.HandleFunc("POST /api/v1/negocios/imoveis/{id}/fotos/upload", imovel.UploadFoto)
	mux.HandleFunc("DELETE /api/v1/negocios/imoveis/{id}/fotos/{foto_id}", imovel.DeleteFoto)

	mux.HandleFunc("GET /api/v1/negocios/clientes", cliente.List)
	mux.HandleFunc("POST /api/v1/negocios/clientes", cliente.Create)
	mux.HandleFunc("GET /api/v1/negocios/clientes/{id}", cliente.GetByID)
	mux.HandleFunc("PUT /api/v1/negocios/clientes/{id}", cliente.Update)
	mux.HandleFunc("DELETE /api/v1/negocios/clientes/{id}", cliente.Delete)

	mux.HandleFunc("GET /api/v1/negocios/agendamentos", agendamento.List)
	mux.HandleFunc("POST /api/v1/negocios/agendamentos", agendamento.Create)
	mux.HandleFunc("PUT /api/v1/negocios/agendamentos/{id}", agendamento.Update)
	mux.HandleFunc("DELETE /api/v1/negocios/agendamentos/{id}", agendamento.Delete)

	mux.HandleFunc("GET /api/v1/negocios/clientes/{cliente_id}/interacoes", interacao.ListByCliente)
	mux.HandleFunc("POST /api/v1/negocios/clientes/{cliente_id}/interacoes", interacao.Create)
	mux.HandleFunc("DELETE /api/v1/negocios/interacoes/{id}", interacao.Delete)

	mux.HandleFunc("GET /api/v1/negocios/leads", lead.List)
	mux.HandleFunc("POST /api/v1/negocios/leads/{id}/atribuir", lead.Atribuir)
	mux.HandleFunc("POST /api/v1/negocios/leads/{id}/qualificar", lead.Qualificar)
	mux.HandleFunc("POST /api/v1/negocios/leads/{id}/descartar", lead.Descartar)
}

func responderSaude(w http.ResponseWriter, _ *http.Request) {
	responderJSON(w, http.StatusOK, map[string]string{"status": "ok", "servico": "kaptei-api"})
}

func responderProntidao(banco *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelar := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancelar()
		if err := banco.PingContext(ctx); err != nil {
			responderJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "indisponivel"})
			return
		}
		responderJSON(w, http.StatusOK, map[string]string{"status": "pronto"})
	}
}

func responderUsuarioAtual(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autorizado"})
		return
	}
	responderJSON(w, http.StatusOK, usuario)
}

func responderJSON(w http.ResponseWriter, status int, dados interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dados)
}
