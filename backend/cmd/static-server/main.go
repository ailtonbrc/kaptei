package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	port := flag.String("port", "8013", "Porta do servidor estático")
	dir := flag.String("dir", "./dist", "Diretório para servir os arquivos")
	apiBase := flag.String("api", "http://127.0.0.1:8080/api", "URL base da API para pré-renderização")
	urlPublica := flag.String("public-url", "", "URL pública canônica; vazio usa o host da requisição")
	flag.Parse()

	// Resolve o caminho absoluto do diretório
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Erro ao ler diretório: %v", err)
	}

	// Verifica se o diretório existe
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		log.Fatalf("O diretório %s não existe", absDir)
	}

	preRenderizador, err := novoPreRenderizador(absDir, *apiBase, *urlPublica)
	if err != nil {
		log.Fatalf("Erro ao preparar SEO server-side: %v", err)
	}
	// Manipulador para SPA (Single Page Application)
	fileServer := http.FileServer(http.Dir(absDir))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
			return
		}
		// Resolve o caminho físico do arquivo
		caminhoRelativo := strings.TrimPrefix(filepath.Clean("/"+r.URL.Path), string(filepath.Separator))
		caminho := filepath.Join(absDir, caminhoRelativo)
		info, err := os.Stat(caminho)
		if preRenderizador.Tentar(w, r) {
			return
		}

		// Se o arquivo não existir ou for um diretório (e não o index), serve o index.html da raiz (SPA)
		if os.IsNotExist(err) || (err == nil && info.IsDir()) {
			http.ServeFile(w, r, filepath.Join(absDir, "index.html"))
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	fmt.Printf("Servidor Estático SPA rodando na porta %s servindo %s...\n", *port, absDir)
	servidor := &http.Server{
		Addr:              ":" + *port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	erros := make(chan error, 1)
	go func() { erros <- servidor.ListenAndServe() }()
	sinais := make(chan os.Signal, 1)
	signal.Notify(sinais, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err = <-erros:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	case <-sinais:
		ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		if err := servidor.Shutdown(ctx); err != nil {
			log.Printf("Erro ao encerrar o servidor estático: %v", err)
		}
	}
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
