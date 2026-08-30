package observabilidade

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/msdev/kaptei/internal/core/ports"
)

const chaveConfiguracao = "OBSERVABILIDADE_CONFIG"

type Metricas struct {
	requisicoes       *prometheus.CounterVec
	duracaoRequisicao *prometheus.HistogramVec
	itensFila         *prometheus.CounterVec
	duracaoFila       *prometheus.HistogramVec
}

func NovoMetricas(db *sql.DB, configuracoes ports.ConfiguracaoRepository, protetor ports.ProtetorSegredos) (*Metricas, http.Handler, error) {
	registro := prometheus.NewRegistry()
	metricas := &Metricas{
		requisicoes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "kaptei", Subsystem: "http", Name: "requisicoes_total",
			Help: "Total de requisições HTTP processadas.",
		}, []string{"metodo", "rota", "status"}),
		duracaoRequisicao: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "kaptei", Subsystem: "http", Name: "duracao_segundos",
			Help:    "Duração das requisições HTTP em segundos.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"metodo", "rota", "status"}),
		itensFila: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "kaptei", Subsystem: "filas", Name: "itens_total",
			Help: "Total de itens processados por fila e resultado.",
		}, []string{"fila", "resultado"}),
		duracaoFila: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "kaptei", Subsystem: "filas", Name: "duracao_processamento_segundos",
			Help:    "Duração do processamento de itens de fila em segundos.",
			Buckets: prometheus.DefBuckets,
		}, []string{"fila", "resultado"}),
	}
	if err := registro.Register(metricas.requisicoes); err != nil {
		return nil, nil, err
	}
	registro.MustRegister(metricas.duracaoRequisicao, metricas.itensFila, metricas.duracaoFila)
	registro.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registro.MustRegister(novoColetorOperacional(db))

	handler := &handlerProtegido{
		proximo:       promhttp.HandlerFor(registro, promhttp.HandlerOpts{EnableOpenMetrics: true}),
		configuracoes: configuracoes, protetor: protetor,
	}
	return metricas, handler, nil
}

func (m *Metricas) RegistrarHTTP(_ context.Context, metodo, rota string, status int, duracao time.Duration) {
	statusTexto := strconv.Itoa(status)
	m.requisicoes.WithLabelValues(metodo, rota, statusTexto).Inc()
	m.duracaoRequisicao.WithLabelValues(metodo, rota, statusTexto).Observe(duracao.Seconds())
}

func (m *Metricas) RegistrarItemFila(_ context.Context, fila, resultado string, duracao time.Duration) {
	m.itensFila.WithLabelValues(fila, resultado).Inc()
	m.duracaoFila.WithLabelValues(fila, resultado).Observe(duracao.Seconds())
}

type configuracaoAcesso struct {
	Ativa bool   `json:"ativa"`
	Token string `json:"token"`
}

type handlerProtegido struct {
	proximo       http.Handler
	configuracoes ports.ConfiguracaoRepository
	protetor      ports.ProtetorSegredos
	mutex         sync.Mutex
	configuracao  configuracaoAcesso
	validaAte     time.Time
}

func (h *handlerProtegido) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	configuracao, err := h.obterConfiguracao(r.Context())
	if err != nil || !configuracao.Ativa {
		http.NotFound(w, r)
		return
	}
	fornecido := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if fornecido == "" || subtle.ConstantTimeCompare([]byte(fornecido), []byte(configuracao.Token)) != 1 {
		w.Header().Set("WWW-Authenticate", `Bearer realm="metricas"`)
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}
	h.proximo.ServeHTTP(w, r)
}

func (h *handlerProtegido) obterConfiguracao(ctx context.Context) (configuracaoAcesso, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if time.Now().Before(h.validaAte) {
		return h.configuracao, nil
	}
	registro, err := h.configuracoes.Get(ctx, chaveConfiguracao)
	if err != nil || registro == nil {
		return configuracaoAcesso{}, err
	}
	var configuracao configuracaoAcesso
	if err := json.Unmarshal(registro.Valor, &configuracao); err != nil {
		return configuracaoAcesso{}, fmt.Errorf("decodificar configuração de métricas: %w", err)
	}
	if configuracao.Ativa {
		configuracao.Token, err = h.protetor.Revelar(configuracao.Token)
		if err != nil {
			return configuracaoAcesso{}, fmt.Errorf("revelar token de métricas: %w", err)
		}
	}
	h.configuracao = configuracao
	h.validaAte = time.Now().Add(30 * time.Second)
	return configuracao, nil
}

type coletorOperacional struct {
	db              *sql.DB
	conexoes        *prometheus.Desc
	backlog         *prometheus.Desc
	idadeMaisAntigo *prometheus.Desc
}

func novoColetorOperacional(db *sql.DB) prometheus.Collector {
	return &coletorOperacional{
		db:              db,
		conexoes:        prometheus.NewDesc("kaptei_banco_conexoes", "Conexões do pool PostgreSQL por estado.", []string{"estado"}, nil),
		backlog:         prometheus.NewDesc("kaptei_filas_backlog", "Quantidade atual de eventos por fila e status.", []string{"fila", "status"}, nil),
		idadeMaisAntigo: prometheus.NewDesc("kaptei_filas_evento_mais_antigo_segundos", "Idade do evento mais antigo por fila e status.", []string{"fila", "status"}, nil),
	}
}

func (c *coletorOperacional) Describe(canal chan<- *prometheus.Desc) {
	canal <- c.conexoes
	canal <- c.backlog
	canal <- c.idadeMaisAntigo
}

func (c *coletorOperacional) Collect(canal chan<- prometheus.Metric) {
	estatisticas := c.db.Stats()
	canal <- prometheus.MustNewConstMetric(c.conexoes, prometheus.GaugeValue, float64(estatisticas.OpenConnections), "abertas")
	canal <- prometheus.MustNewConstMetric(c.conexoes, prometheus.GaugeValue, float64(estatisticas.InUse), "em_uso")
	canal <- prometheus.MustNewConstMetric(c.conexoes, prometheus.GaugeValue, float64(estatisticas.Idle), "ociosas")

	ctx, cancelar := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelar()
	rows, err := c.db.QueryContext(ctx, `
		SELECT fila,status,total,idade_segundos FROM (
			SELECT 'outbox'::text AS fila,status,COUNT(*)::float8 AS total,
				EXTRACT(EPOCH FROM NOW()-MIN(criado_em))::float8 AS idade_segundos
			FROM eventos_outbox GROUP BY status
			UNION ALL
			SELECT LOWER(provedor)::text AS fila,status,COUNT(*)::float8 AS total,
				EXTRACT(EPOCH FROM NOW()-MIN(criado_em))::float8 AS idade_segundos
			FROM eventos_integracao GROUP BY provedor,status
		) dados`)
	if err != nil {
		canal <- prometheus.NewInvalidMetric(c.backlog, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var fila, status string
		var total, idade float64
		if err := rows.Scan(&fila, &status, &total, &idade); err != nil {
			canal <- prometheus.NewInvalidMetric(c.backlog, err)
			return
		}
		canal <- prometheus.MustNewConstMetric(c.backlog, prometheus.GaugeValue, total, fila, status)
		canal <- prometheus.MustNewConstMetric(c.idadeMaisAntigo, prometheus.GaugeValue, idade, fila, status)
	}
	if err := rows.Err(); err != nil && !errors.Is(err, context.Canceled) {
		canal <- prometheus.NewInvalidMetric(c.backlog, err)
	}
}
