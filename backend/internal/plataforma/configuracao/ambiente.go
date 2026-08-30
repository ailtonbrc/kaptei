package configuracao

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const tamanhoMinimoSegredoJWT = 32

type Ambiente struct {
	DatabaseURL                       string
	JWTSecret                         string
	PortaHTTP                         string
	OrigensCORS                       []string
	PicPayToken                       string
	PicPaySellerToken                 string
	TimeoutLeitura                    time.Duration
	TimeoutEscrita                    time.Duration
	TimeoutOcioso                     time.Duration
	LimiteLeadsPublicos               int
	JanelaLeadsPublicos               time.Duration
	ConfiarProxy                      bool
	CookieSeguro                      bool
	LimiteAutenticacao                int
	JanelaAutenticacao                time.Duration
	LimiteLeituraPublica              int
	JanelaLeituraPublica              time.Duration
	StripeSecretKey                   string
	StripeWebhookSecret               string
	URLPublicaAplicacao               string
	ChaveCriptografia                 string
	MaximoCabecalhoBytes              int
	BancoMaximoAbertas                int
	BancoMaximoOciosas                int
	BancoVidaMaxima                   time.Duration
	BancoOciosidadeMaxima             time.Duration
	OutboxIntervalo                   time.Duration
	OutboxTamanhoLote                 int
	OutboxMaxTentativas               int
	OutboxDuracaoBloqueio             time.Duration
	OutboxBackoffInicial              time.Duration
	OutboxBackoffMaximo               time.Duration
	StorageProvider                   string
	StorageLocalDir                   string
	StoragePublicBaseURL              string
	StorageS3Region                   string
	StorageS3Bucket                   string
	StorageS3Endpoint                 string
	StorageS3AccessKey                string
	StorageS3SecretKey                string
	MetricasConversaoExpurgoIntervalo time.Duration
	StorageS3PathStyle                bool
	ImagemMaximoBytes                 int
	ImagemMaximoPixels                int64
	ImagemMaximoPrincipal             int
	ImagemMaximoThumbnail             int
	ImagemQualidadeJPEG               int
	ImagemMaximoConcorrente           int
	MetaAppSecret                     string
	MetaWebhookVerifyToken            string
	MetaGraphBaseURL                  string
	GrupoOLXWebhookSecret             string
	MetaGraphAPIVersion               string
	MetaHTTPTimeout                   time.Duration
}

func Carregar() (Ambiente, error) {
	carregarArquivosLocais()
	if err := validarVariaveisEscalares(); err != nil {
		return Ambiente{}, err
	}
	databaseURL, err := obterDatabaseURLDoAmbiente()
	if err != nil {
		return Ambiente{}, err
	}
	portaHTTP := valorOuPadrao("PORT", "8080")
	urlPublicaPadrao := ""
	storagePublicoPadrao := ""
	if ambienteLocal() {
		urlPublicaPadrao = "http://localhost:5173"
		storagePublicoPadrao = "http://localhost:" + portaHTTP + "/arquivos"
	}

	ambiente := Ambiente{
		DatabaseURL:                       databaseURL,
		JWTSecret:                         os.Getenv("JWT_SECRET"),
		PortaHTTP:                         portaHTTP,
		OrigensCORS:                       listaOuPadrao("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		PicPayToken:                       strings.TrimSpace(os.Getenv("PICPAY_TOKEN")),
		PicPaySellerToken:                 strings.TrimSpace(os.Getenv("PICPAY_SELLER_TOKEN")),
		TimeoutLeitura:                    duracaoOuPadrao("HTTP_READ_TIMEOUT", 15*time.Second),
		TimeoutEscrita:                    duracaoOuPadrao("HTTP_WRITE_TIMEOUT", 30*time.Second),
		TimeoutOcioso:                     duracaoOuPadrao("HTTP_IDLE_TIMEOUT", 60*time.Second),
		LimiteLeadsPublicos:               inteiroOuPadrao("PUBLIC_LEAD_RATE_LIMIT", 10),
		JanelaLeadsPublicos:               duracaoOuPadrao("PUBLIC_LEAD_RATE_WINDOW", time.Minute),
		ConfiarProxy:                      strings.EqualFold(strings.TrimSpace(os.Getenv("TRUST_PROXY_HEADERS")), "true"),
		CookieSeguro:                      strings.EqualFold(strings.TrimSpace(os.Getenv("COOKIE_SECURE")), "true"),
		LimiteAutenticacao:                inteiroOuPadrao("AUTH_RATE_LIMIT", 10),
		JanelaAutenticacao:                duracaoOuPadrao("AUTH_RATE_WINDOW", 5*time.Minute),
		LimiteLeituraPublica:              inteiroOuPadrao("PUBLIC_READ_RATE_LIMIT", 180),
		JanelaLeituraPublica:              duracaoOuPadrao("PUBLIC_READ_RATE_WINDOW", time.Minute),
		StripeSecretKey:                   strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")),
		StripeWebhookSecret:               strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET")),
		URLPublicaAplicacao:               strings.TrimRight(valorOuPadrao("APP_PUBLIC_URL", urlPublicaPadrao), "/"),
		ChaveCriptografia:                 strings.TrimSpace(os.Getenv("CONFIG_ENCRYPTION_KEY")),
		MaximoCabecalhoBytes:              inteiroOuPadrao("HTTP_MAX_HEADER_BYTES", 1<<20),
		BancoMaximoAbertas:                inteiroOuPadrao("DB_MAX_OPEN_CONNS", 25),
		BancoMaximoOciosas:                inteiroOuPadrao("DB_MAX_IDLE_CONNS", 10),
		BancoVidaMaxima:                   duracaoOuPadrao("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		BancoOciosidadeMaxima:             duracaoOuPadrao("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		OutboxIntervalo:                   duracaoOuPadrao("OUTBOX_POLL_INTERVAL", 2*time.Second),
		OutboxTamanhoLote:                 inteiroOuPadrao("OUTBOX_BATCH_SIZE", 20),
		OutboxMaxTentativas:               inteiroOuPadrao("OUTBOX_MAX_ATTEMPTS", 8),
		OutboxDuracaoBloqueio:             duracaoOuPadrao("OUTBOX_LOCK_DURATION", time.Minute),
		OutboxBackoffInicial:              duracaoOuPadrao("OUTBOX_INITIAL_BACKOFF", 30*time.Second),
		OutboxBackoffMaximo:               duracaoOuPadrao("OUTBOX_MAX_BACKOFF", 30*time.Minute),
		StorageProvider:                   strings.ToLower(valorOuPadrao("STORAGE_PROVIDER", "local")),
		StorageLocalDir:                   valorOuPadrao("STORAGE_LOCAL_DIR", "./data/objetos"),
		StoragePublicBaseURL:              strings.TrimRight(valorOuPadrao("STORAGE_PUBLIC_BASE_URL", storagePublicoPadrao), "/"),
		StorageS3Region:                   strings.TrimSpace(os.Getenv("STORAGE_S3_REGION")),
		StorageS3Bucket:                   strings.TrimSpace(os.Getenv("STORAGE_S3_BUCKET")),
		MetricasConversaoExpurgoIntervalo: duracaoOuPadrao("METRICAS_CONVERSAO_EXPURGO_INTERVAL", 24*time.Hour),
		StorageS3Endpoint:                 strings.TrimRight(strings.TrimSpace(os.Getenv("STORAGE_S3_ENDPOINT")), "/"),
		StorageS3AccessKey:                strings.TrimSpace(os.Getenv("STORAGE_S3_ACCESS_KEY")),
		StorageS3SecretKey:                strings.TrimSpace(os.Getenv("STORAGE_S3_SECRET_KEY")),
		StorageS3PathStyle:                strings.EqualFold(strings.TrimSpace(os.Getenv("STORAGE_S3_PATH_STYLE")), "true"),
		ImagemMaximoBytes:                 inteiroOuPadrao("IMAGE_MAX_BYTES", 15*1024*1024),
		ImagemMaximoPixels:                int64(inteiroOuPadrao("IMAGE_MAX_PIXELS", 24_000_000)),
		ImagemMaximoPrincipal:             inteiroOuPadrao("IMAGE_MAIN_MAX_DIMENSION", 2400),
		ImagemMaximoThumbnail:             inteiroOuPadrao("IMAGE_THUMB_MAX_DIMENSION", 640),
		ImagemQualidadeJPEG:               inteiroOuPadrao("IMAGE_JPEG_QUALITY", 85),
		ImagemMaximoConcorrente:           inteiroOuPadrao("IMAGE_MAX_CONCURRENT_PROCESSING", 2),
		MetaAppSecret:                     strings.TrimSpace(os.Getenv("META_APP_SECRET")),
		MetaWebhookVerifyToken:            strings.TrimSpace(os.Getenv("META_WEBHOOK_VERIFY_TOKEN")),
		MetaGraphBaseURL:                  strings.TrimRight(strings.TrimSpace(os.Getenv("META_GRAPH_BASE_URL")), "/"),
		GrupoOLXWebhookSecret:             strings.TrimSpace(os.Getenv("GRUPO_OLX_WEBHOOK_SECRET")),
		MetaGraphAPIVersion:               strings.TrimSpace(os.Getenv("META_GRAPH_API_VERSION")),
		MetaHTTPTimeout:                   duracaoOuPadrao("META_HTTP_TIMEOUT", 15*time.Second),
	}

	if ambiente.DatabaseURL == "" {
		return Ambiente{}, errors.New("DATABASE_URL não configurada")
	}
	if len(ambiente.JWTSecret) < tamanhoMinimoSegredoJWT {
		return Ambiente{}, errors.New("JWT_SECRET deve possuir pelo menos 32 caracteres")
	}
	chaveCriptografia, err := base64.StdEncoding.DecodeString(ambiente.ChaveCriptografia)
	if err != nil || len(chaveCriptografia) != 32 {
		return Ambiente{}, errors.New("CONFIG_ENCRYPTION_KEY deve possuir 32 bytes aleatórios em Base64")
	}
	if err := validarOrigensCORS(ambiente.OrigensCORS); err != nil {
		return Ambiente{}, err
	}
	if ambiente.BancoMaximoOciosas > ambiente.BancoMaximoAbertas {
		return Ambiente{}, errors.New("DB_MAX_IDLE_CONNS não pode exceder DB_MAX_OPEN_CONNS")
	}
	if ambiente.URLPublicaAplicacao == "" {
		return Ambiente{}, errors.New("APP_PUBLIC_URL não configurada")
	}
	if ambiente.OutboxBackoffInicial > ambiente.OutboxBackoffMaximo {
		return Ambiente{}, errors.New("OUTBOX_INITIAL_BACKOFF não pode exceder OUTBOX_MAX_BACKOFF")
	}
	if ambiente.StorageProvider != "local" && ambiente.StorageProvider != "s3" {
		return Ambiente{}, errors.New("STORAGE_PROVIDER deve ser local ou s3")
	}
	if ambiente.StoragePublicBaseURL == "" {
		return Ambiente{}, errors.New("STORAGE_PUBLIC_BASE_URL não configurada")
	}
	if err := validarURLPublica(ambiente.StoragePublicBaseURL); err != nil {
		return Ambiente{}, fmt.Errorf("STORAGE_PUBLIC_BASE_URL inválida: %w", err)
	}
	if ambiente.StorageProvider == "s3" && (ambiente.StorageS3Region == "" || ambiente.StorageS3Bucket == "") {
		return Ambiente{}, errors.New("STORAGE_S3_REGION e STORAGE_S3_BUCKET são obrigatórios para S3")
	}
	if (ambiente.StorageS3AccessKey == "") != (ambiente.StorageS3SecretKey == "") {
		return Ambiente{}, errors.New("STORAGE_S3_ACCESS_KEY e STORAGE_S3_SECRET_KEY devem ser informadas em conjunto")
	}
	if ambiente.StorageS3Endpoint != "" {
		if err := validarURLPublica(ambiente.StorageS3Endpoint); err != nil {
			return Ambiente{}, fmt.Errorf("STORAGE_S3_ENDPOINT inválido: %w", err)
		}
	}
	if ambiente.ImagemMaximoThumbnail > ambiente.ImagemMaximoPrincipal {
		return Ambiente{}, errors.New("IMAGE_THUMB_MAX_DIMENSION não pode exceder IMAGE_MAIN_MAX_DIMENSION")
	}
	if err := validarURLPublica(ambiente.URLPublicaAplicacao); err != nil {
		return Ambiente{}, fmt.Errorf("APP_PUBLIC_URL inválida: %w", err)
	}
	if ambiente.MetaHabilitada() {
		if ambiente.MetaAppSecret == "" || ambiente.MetaWebhookVerifyToken == "" || ambiente.MetaGraphBaseURL == "" || ambiente.MetaGraphAPIVersion == "" {
			return Ambiente{}, errors.New("META_APP_SECRET, META_WEBHOOK_VERIFY_TOKEN, META_GRAPH_BASE_URL e META_GRAPH_API_VERSION devem ser configuradas em conjunto")
		}
		if len(ambiente.MetaAppSecret) < 16 || len(ambiente.MetaWebhookVerifyToken) < 32 {
			return Ambiente{}, errors.New("segredos da integraÃ§Ã£o Meta sÃ£o muito curtos")
		}
		if err := validarURLHTTPS(ambiente.MetaGraphBaseURL); err != nil {
			return Ambiente{}, fmt.Errorf("META_GRAPH_BASE_URL invÃ¡lida: %w", err)
		}
	}

	if ambiente.GrupoOLXWebhookSecret != "" && len(ambiente.GrupoOLXWebhookSecret) < 16 {
		return Ambiente{}, errors.New("GRUPO_OLX_WEBHOOK_SECRET deve possuir pelo menos 16 caracteres")
	}
	return ambiente, nil
}

func validarVariaveisEscalares() error {
	inteiros := map[string][2]int{
		"PORT": {1, 65535}, "AUTH_RATE_LIMIT": {1, 1_000_000}, "PUBLIC_LEAD_RATE_LIMIT": {1, 1_000_000},
		"PUBLIC_READ_RATE_LIMIT": {1, 1_000_000}, "HTTP_MAX_HEADER_BYTES": {16 << 10, 8 << 20},
		"DB_PORT":           {1, 65535},
		"DB_MAX_OPEN_CONNS": {1, 1000}, "DB_MAX_IDLE_CONNS": {1, 1000},
		"OUTBOX_BATCH_SIZE": {1, 1000}, "OUTBOX_MAX_ATTEMPTS": {1, 100},
		"IMAGE_MAX_BYTES": {64 * 1024, 50 * 1024 * 1024}, "IMAGE_MAX_PIXELS": {1_000_000, 100_000_000},
		"IMAGE_MAIN_MAX_DIMENSION": {640, 8000}, "IMAGE_THUMB_MAX_DIMENSION": {128, 2000},
		"IMAGE_JPEG_QUALITY":              {50, 95},
		"IMAGE_MAX_CONCURRENT_PROCESSING": {1, 32},
	}
	for chave, intervalo := range inteiros {
		valor := strings.TrimSpace(os.Getenv(chave))
		if valor == "" {
			continue
		}
		numero, err := strconv.Atoi(valor)
		if err != nil || numero < intervalo[0] || numero > intervalo[1] {
			return fmt.Errorf("%s deve ser um inteiro entre %d e %d", chave, intervalo[0], intervalo[1])
		}
	}
	duracoes := []string{"HTTP_READ_TIMEOUT", "HTTP_WRITE_TIMEOUT", "HTTP_IDLE_TIMEOUT", "AUTH_RATE_WINDOW", "PUBLIC_LEAD_RATE_WINDOW", "PUBLIC_READ_RATE_WINDOW", "DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_IDLE_TIME", "OUTBOX_POLL_INTERVAL", "OUTBOX_LOCK_DURATION", "OUTBOX_INITIAL_BACKOFF", "OUTBOX_MAX_BACKOFF", "META_HTTP_TIMEOUT", "METRICAS_CONVERSAO_EXPURGO_INTERVAL"}
	for _, chave := range duracoes {
		valor := strings.TrimSpace(os.Getenv(chave))
		if valor == "" {
			continue
		}
		duracao, err := time.ParseDuration(valor)
		if err != nil || duracao <= 0 || duracao > 24*time.Hour {
			return fmt.Errorf("%s deve ser uma duração positiva de até 24h", chave)
		}
	}
	return nil
}

func validarOrigensCORS(origens []string) error {
	for _, origem := range origens {
		if strings.Contains(origem, "*") {
			return errors.New("CORS_ALLOWED_ORIGINS não pode conter curingas quando credenciais estão habilitadas")
		}
		endereco, err := url.Parse(origem)
		if err != nil || endereco.Scheme == "" || endereco.Host == "" || endereco.User != nil ||
			(endereco.Scheme != "http" && endereco.Scheme != "https") ||
			(endereco.Path != "" && endereco.Path != "/") || endereco.RawQuery != "" || endereco.Fragment != "" {
			return fmt.Errorf("origem CORS inválida: %q", origem)
		}
	}
	return nil
}

func validarURLPublica(valor string) error {
	endereco, err := url.Parse(valor)
	if err != nil || endereco.Host == "" || endereco.User != nil ||
		(endereco.Scheme != "http" && endereco.Scheme != "https") || endereco.RawQuery != "" || endereco.Fragment != "" {
		return errors.New("use uma URL absoluta HTTP ou HTTPS sem credenciais, consulta ou fragmento")
	}
	return nil
}

func validarURLHTTPS(valor string) error {
	endereco, err := url.Parse(valor)
	if err != nil || endereco.Scheme != "https" || endereco.Host == "" || endereco.User != nil ||
		endereco.RawQuery != "" || endereco.Fragment != "" {
		return errors.New("use uma URL HTTPS absoluta sem credenciais, consulta ou fragmento")
	}
	return nil
}

func (a Ambiente) MetaHabilitada() bool {
	return a.MetaAppSecret != "" || a.MetaWebhookVerifyToken != "" || a.MetaGraphBaseURL != "" || a.MetaGraphAPIVersion != ""
}

func inteiroOuPadrao(chave string, padrao int) int {
	valor := strings.TrimSpace(os.Getenv(chave))
	if valor == "" {
		return padrao
	}
	numero, err := strconv.Atoi(valor)
	if err != nil || numero <= 0 {
		return padrao
	}
	return numero
}

func CarregarDatabaseURL() (string, error) {
	carregarArquivosLocais()
	databaseURL, err := obterDatabaseURLDoAmbiente()
	if err != nil {
		return "", err
	}
	if databaseURL == "" {
		return "", errors.New("DATABASE_URL não configurada")
	}
	return databaseURL, nil
}

// obterDatabaseURLDoAmbiente preserva instalações locais antigas sem reduzir
// a exigência de DATABASE_URL nos ambientes novos. A biblioteca padrão escapa
// usuário e senha para que segredos nunca sejam concatenados manualmente.
func obterDatabaseURLDoAmbiente() (string, error) {
	if databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); databaseURL != "" {
		return databaseURL, nil
	}

	legado := map[string]string{
		"DB_HOST":     strings.TrimSpace(os.Getenv("DB_HOST")),
		"DB_USER":     strings.TrimSpace(os.Getenv("DB_USER")),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_DATABASE": strings.TrimSpace(os.Getenv("DB_DATABASE")),
	}
	algumConfigurado := false
	faltantes := make([]string, 0, len(legado))
	for chave, valor := range legado {
		if valor != "" {
			algumConfigurado = true
		} else {
			faltantes = append(faltantes, chave)
		}
	}
	if !algumConfigurado {
		return "", nil
	}
	if len(faltantes) > 0 {
		return "", fmt.Errorf("configuração PostgreSQL legada incompleta; faltam: %s", strings.Join(faltantes, ", "))
	}

	porta := strings.TrimSpace(os.Getenv("DB_PORT"))
	if porta == "" {
		porta = "5432"
	}
	portaNumero, err := strconv.Atoi(porta)
	if err != nil || portaNumero < 1 || portaNumero > 65535 {
		return "", errors.New("DB_PORT deve ser um inteiro entre 1 e 65535")
	}

	modoSSL := strings.ToLower(strings.TrimSpace(os.Getenv("DB_SSLMODE")))
	if modoSSL == "" {
		modoSSL = "require"
		if ambienteLocal() {
			modoSSL = "disable"
		}
	}
	modosPermitidos := map[string]bool{
		"disable": true, "allow": true, "prefer": true, "require": true,
		"verify-ca": true, "verify-full": true,
	}
	if !modosPermitidos[modoSSL] {
		return "", errors.New("DB_SSLMODE inválido")
	}

	endereco := &url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(legado["DB_HOST"], porta),
		Path:   "/" + legado["DB_DATABASE"],
		User:   url.UserPassword(legado["DB_USER"], legado["DB_PASSWORD"]),
	}
	consulta := endereco.Query()
	consulta.Set("sslmode", modoSSL)
	endereco.RawQuery = consulta.Encode()
	return endereco.String(), nil
}

func carregarArquivosLocais() {
	if executavel, err := os.Executable(); err == nil {
		_ = godotenv.Load(filepath.Join(filepath.Dir(executavel), ".env"))
	}
	_ = godotenv.Load(".env")
}

func valorOuPadrao(chave, padrao string) string {
	if valor := strings.TrimSpace(os.Getenv(chave)); valor != "" {
		return valor
	}
	return padrao
}

func listaOuPadrao(chave string, padrao []string) []string {
	valor := strings.TrimSpace(os.Getenv(chave))
	if valor == "" {
		return padrao
	}

	itens := make([]string, 0)
	for _, item := range strings.Split(valor, ",") {
		if origem := strings.TrimSpace(item); origem != "" {
			itens = append(itens, origem)
		}
	}
	if len(itens) == 0 {
		return padrao
	}
	return itens
}

func duracaoOuPadrao(chave string, padrao time.Duration) time.Duration {
	valor := strings.TrimSpace(os.Getenv(chave))
	if valor == "" {
		return padrao
	}
	duracao, err := time.ParseDuration(valor)
	if err != nil || duracao <= 0 {
		return padrao
	}
	return duracao
}

func ambienteLocal() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ENV"))) {
	case "development", "desenvolvimento", "local", "dev":
		return true
	default:
		return false
	}
}
