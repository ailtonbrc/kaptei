package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/msdev/kaptei/internal/adapters/gateways"
	"github.com/msdev/kaptei/internal/adapters/repositories"
	servicoemail "github.com/msdev/kaptei/internal/adapters/services/email"
	servicowhatsapp "github.com/msdev/kaptei/internal/adapters/services/whatsapp"
	"github.com/msdev/kaptei/internal/adapters/storage"
	"github.com/msdev/kaptei/internal/core/ports"
	servicoscore "github.com/msdev/kaptei/internal/core/services"
	"github.com/msdev/kaptei/internal/plataforma/bancodados"
	"github.com/msdev/kaptei/internal/plataforma/configuracao"
	"github.com/msdev/kaptei/internal/plataforma/httpapi"
	"github.com/msdev/kaptei/internal/plataforma/observabilidade"
	"github.com/msdev/kaptei/internal/plataforma/seguranca"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ambiente, err := configuracao.Carregar()
	if err != nil {
		slog.Error("configuração inválida", "erro", err)
		os.Exit(1)
	}
	banco, err := bancodados.AbrirPostgresConfigurado(ambiente.DatabaseURL, bancodados.ConfiguracaoPool{
		MaximoAbertas: ambiente.BancoMaximoAbertas, MaximoOciosas: ambiente.BancoMaximoOciosas,
		VidaMaxima: ambiente.BancoVidaMaxima, OciosidadeMaxima: ambiente.BancoOciosidadeMaxima,
	})
	if err != nil {
		slog.Error("falha ao abrir banco de dados", "erro", err)
		os.Exit(1)
	}
	defer banco.Close()

	protetorSegredos, err := seguranca.NovoProtetorSegredos(ambiente.ChaveCriptografia)
	if err != nil {
		slog.Error("falha ao configurar proteção de segredos", "erro", err)
		os.Exit(1)
	}
	codecEmail := servicoemail.NewCodecOutbox(protetorSegredos, ambiente.OutboxMaxTentativas)
	codecObjeto := storage.NewCodecObjetoOutbox(protetorSegredos, ambiente.OutboxMaxTentativas)
	codecWhatsApp := servicowhatsapp.NewCodecIntegracao(protetorSegredos, ambiente.OutboxMaxTentativas)
	codecWhatsAppOutbox := servicowhatsapp.NewCodecOutbox(protetorSegredos, ambiente.OutboxMaxTentativas)
	objetos, handlerMidia, prefixoMidia, err := criarArmazenamento(context.Background(), ambiente)
	if err != nil {
		slog.Error("falha ao configurar armazenamento de objetos", "erro", err)
		os.Exit(1)
	}
	processadorImagem := storage.NewProcessadorImagem(
		ambiente.ImagemMaximoBytes, ambiente.ImagemMaximoPixels,
		ambiente.ImagemMaximoPrincipal, ambiente.ImagemMaximoThumbnail, ambiente.ImagemQualidadeJPEG,
	)
	configRepo := repositories.NewConfiguracaoRepository(banco)
	metricas, handlerMetricas, err := observabilidade.NovoMetricas(banco, configRepo, protetorSegredos)
	if err != nil {
		slog.Error("falha ao configurar métricas", "erro", err)
		os.Exit(1)
	}
	entregadorEmail := servicoemail.NewSMTPService(configRepo, protetorSegredos)
	whatsAppRepo := repositories.NewIntegracaoWhatsAppPostgres(banco)
	metricasConversaoRepo := repositories.NewMetricasConversaoPostgres(banco)
	tratadoresOutbox := []ports.TratadorEventoOutbox{
		servicoscore.NewTratadorEmailOutbox(codecEmail, entregadorEmail),
		servicoscore.NewTratadorObjetoOutbox(codecObjeto, objetos),
	}
	var clienteWhatsApp ports.ClienteWhatsApp
	if ambiente.MetaHabilitada() {
		clienteWhatsApp, err = gateways.NewClienteWhatsAppGraph(ambiente.MetaGraphBaseURL, ambiente.MetaGraphAPIVersion, ambiente.MetaAppSecret, ambiente.MetaHTTPTimeout)
		if err != nil {
			slog.Error("falha ao configurar WhatsApp Graph", "erro", err)
			os.Exit(1)
		}
	}
	tratadoresOutbox = append(tratadoresOutbox, servicoscore.NewTratadorWhatsAppOutbox(codecWhatsAppOutbox, clienteWhatsApp, whatsAppRepo, protetorSegredos))
	processadorOutbox := servicoscore.NewProcessadorOutbox(
		repositories.NewOutboxPostgres(banco),
		tratadoresOutbox,
		servicoscore.ConfiguracaoProcessadorOutbox{
			TrabalhadorID:   identificadorTrabalhador(),
			Intervalo:       ambiente.OutboxIntervalo,
			TamanhoLote:     ambiente.OutboxTamanhoLote,
			DuracaoBloqueio: ambiente.OutboxDuracaoBloqueio,
			BackoffInicial:  ambiente.OutboxBackoffInicial,
			BackoffMaximo:   ambiente.OutboxBackoffMaximo,
			Metricas:        metricas,
		},
	)
	contextoAplicacao, cancelarAplicacao := context.WithCancel(context.Background())
	processadorExpurgoMetricasConversao := servicoscore.NewProcessadorExpurgoMetricasConversao(metricasConversaoRepo, ambiente.MetricasConversaoExpurgoIntervalo)

	var trabalhadores sync.WaitGroup
	trabalhadores.Add(1)
	go func() {
		defer trabalhadores.Done()
		processadorOutbox.Executar(contextoAplicacao)
	}()

	trabalhadores.Add(1)
	go func() {
		defer trabalhadores.Done()
		processadorExpurgoMetricasConversao.Executar(contextoAplicacao)
	}()
	if ambiente.MetaHabilitada() {
		clienteMeta, erroMeta := gateways.NewClienteMetaGraph(ambiente.MetaGraphBaseURL, ambiente.MetaGraphAPIVersion, ambiente.MetaAppSecret, ambiente.MetaHTTPTimeout)
		if erroMeta != nil {
			slog.Error("falha ao configurar Meta Graph", "erro", erroMeta)
			os.Exit(1)
		}
		metaRepo := repositories.NewIntegracaoMetaPostgres(banco)
		leadSvc := servicoscore.NewLeadService(repositories.NewLeadPostgres(banco), repositories.NewContaRepository(banco), repositories.NewUsuarioRepository(banco))
		processadorMeta := servicoscore.NewProcessadorIntegracaoMeta(metaRepo, clienteMeta, protetorSegredos, leadSvc, servicoscore.ConfiguracaoProcessadorOutbox{
			TrabalhadorID: identificadorTrabalhador() + ":meta", Intervalo: ambiente.OutboxIntervalo,
			TamanhoLote: ambiente.OutboxTamanhoLote, DuracaoBloqueio: ambiente.OutboxDuracaoBloqueio,
			BackoffInicial: ambiente.OutboxBackoffInicial, BackoffMaximo: ambiente.OutboxBackoffMaximo,
			Metricas: metricas,
		})
		trabalhadores.Add(1)
		go func() {
			defer trabalhadores.Done()
			processadorMeta.Executar(contextoAplicacao)
		}()
		processadorWhatsApp := servicoscore.NewProcessadorIntegracaoWhatsApp(whatsAppRepo, codecWhatsApp, leadSvc, servicoscore.ConfiguracaoProcessadorOutbox{
			TrabalhadorID: identificadorTrabalhador() + ":whatsapp", Intervalo: ambiente.OutboxIntervalo,
			TamanhoLote: ambiente.OutboxTamanhoLote, DuracaoBloqueio: ambiente.OutboxDuracaoBloqueio,
			BackoffInicial: ambiente.OutboxBackoffInicial, BackoffMaximo: ambiente.OutboxBackoffMaximo,
			Metricas: metricas,
		})
		trabalhadores.Add(1)
		go func() { defer trabalhadores.Done(); processadorWhatsApp.Executar(contextoAplicacao) }()
	}

	servidor := &http.Server{Addr: ":" + ambiente.PortaHTTP, Handler: httpapi.NovoRoteador(banco, ambiente, httpapi.DependenciasRoteador{
		ProtetorSegredos: protetorSegredos, PreparadorEmail: codecEmail,
		ProcessadorImagem: processadorImagem, Objetos: objetos, PreparadorObjeto: codecObjeto,
		PreparadorWhatsApp:       codecWhatsApp,
		PreparadorWhatsAppOutbox: codecWhatsAppOutbox,
		Metricas:                 metricas,
		HandlerMetricas:          handlerMetricas,
		PrefixoMidiaLocal:        prefixoMidia, HandlerMidiaLocal: handlerMidia,
	}),
		ReadTimeout: ambiente.TimeoutLeitura, ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: ambiente.TimeoutEscrita, IdleTimeout: ambiente.TimeoutOcioso, MaxHeaderBytes: ambiente.MaximoCabecalhoBytes}
	erros := make(chan error, 1)
	go func() {
		slog.Info("Kaptei API iniciada", "porta", ambiente.PortaHTTP)
		erros <- servidor.ListenAndServe()
	}()

	sinais := make(chan os.Signal, 1)
	signal.Notify(sinais, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sinal := <-sinais:
		slog.Info("encerrando API", "sinal", sinal.String())
	case err := <-erros:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("falha no servidor HTTP", "erro", err)
		}
	}
	cancelarAplicacao()
	ctx, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()
	if err := servidor.Shutdown(ctx); err != nil {
		slog.Error("falha no encerramento gracioso", "erro", err)
	}
	trabalhadores.Wait()
}

func identificadorTrabalhador() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "kaptei"
	}
	return host + ":" + fmt.Sprint(os.Getpid())
}

func criarArmazenamento(ctx context.Context, ambiente configuracao.Ambiente) (ports.ArmazenamentoObjetos, http.Handler, string, error) {
	if ambiente.StorageProvider == "s3" {
		objetos, err := storage.NewArmazenamentoS3(ctx, storage.ConfiguracaoS3{
			Regiao: ambiente.StorageS3Region, Bucket: ambiente.StorageS3Bucket,
			Endpoint: ambiente.StorageS3Endpoint, AccessKey: ambiente.StorageS3AccessKey,
			SecretKey: ambiente.StorageS3SecretKey, BaseURL: ambiente.StoragePublicBaseURL,
			UsarPathStyle: ambiente.StorageS3PathStyle,
		})
		return objetos, nil, "", err
	}
	objetos, err := storage.NewArmazenamentoLocal(ambiente.StorageLocalDir, ambiente.StoragePublicBaseURL)
	if err != nil {
		return nil, nil, "", err
	}
	return objetos, objetos.HandlerPublico(), objetos.PrefixoPublico(), nil
}
