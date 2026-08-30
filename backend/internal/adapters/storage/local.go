package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ArmazenamentoLocal struct {
	raiz           string
	baseURL        string
	prefixoPublico string
}

func NewArmazenamentoLocal(raiz, baseURL string) (*ArmazenamentoLocal, error) {
	raizAbsoluta, err := filepath.Abs(strings.TrimSpace(raiz))
	if err != nil || raizAbsoluta == "" {
		return nil, errors.New("diretório local de armazenamento inválido")
	}
	endereco, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || endereco.Scheme == "" || endereco.Host == "" {
		return nil, errors.New("URL pública do armazenamento inválida")
	}
	if strings.Trim(endereco.Path, "/") == "" {
		return nil, errors.New("URL pública do armazenamento local deve possuir um prefixo de caminho")
	}
	if err := os.MkdirAll(raizAbsoluta, 0o750); err != nil {
		return nil, fmt.Errorf("preparar armazenamento local: %w", err)
	}
	raizReal, err := filepath.EvalSymlinks(raizAbsoluta)
	if err != nil {
		return nil, fmt.Errorf("resolver armazenamento local: %w", err)
	}
	return &ArmazenamentoLocal{raiz: raizReal, baseURL: endereco.String(), prefixoPublico: strings.TrimRight(endereco.Path, "/") + "/"}, nil
}

func (a *ArmazenamentoLocal) Nome() string { return "local" }

func (a *ArmazenamentoLocal) Salvar(ctx context.Context, chave string, conteudo []byte, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	destino, err := a.resolver(chave)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(destino), 0o750); err != nil {
		return "", fmt.Errorf("preparar diretório do objeto: %w", err)
	}
	if err := a.validarCaminhoReal(filepath.Dir(destino)); err != nil {
		return "", err
	}
	temporario, err := os.CreateTemp(filepath.Dir(destino), ".kaptei-upload-*")
	if err != nil {
		return "", fmt.Errorf("criar arquivo temporário: %w", err)
	}
	nomeTemporario := temporario.Name()
	defer os.Remove(nomeTemporario)
	if err := temporario.Chmod(0o640); err != nil {
		temporario.Close()
		return "", err
	}
	if _, err := temporario.ReadFrom(bytes.NewReader(conteudo)); err != nil {
		temporario.Close()
		return "", fmt.Errorf("gravar objeto: %w", err)
	}
	if err := temporario.Close(); err != nil {
		return "", fmt.Errorf("confirmar objeto temporário: %w", err)
	}
	if err := os.Rename(nomeTemporario, destino); err != nil {
		return "", fmt.Errorf("publicar objeto: %w", err)
	}
	return a.urlPublica(chave), nil
}

func (a *ArmazenamentoLocal) Excluir(ctx context.Context, chave string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	destino, err := a.resolver(chave)
	if err != nil {
		return err
	}
	if _, err := os.Stat(destino); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if err := a.validarCaminhoReal(destino); err != nil {
		return err
	}
	err = os.Remove(destino)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (a *ArmazenamentoLocal) HandlerPublico() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chave := strings.TrimPrefix(r.URL.Path, a.prefixoPublico)
		arquivo, err := a.resolver(chave)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(arquivo)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		if err := a.validarCaminhoReal(arquivo); err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, r, arquivo)
	})
}

func (a *ArmazenamentoLocal) validarCaminhoReal(caminho string) error {
	real, err := filepath.EvalSymlinks(caminho)
	if err != nil {
		return fmt.Errorf("resolver caminho do objeto: %w", err)
	}
	relativo, err := filepath.Rel(a.raiz, real)
	if err != nil || relativo == ".." || strings.HasPrefix(relativo, ".."+string(filepath.Separator)) {
		return errors.New("caminho do objeto sai do armazenamento")
	}
	return nil
}

func (a *ArmazenamentoLocal) PrefixoPublico() string { return a.prefixoPublico }

func (a *ArmazenamentoLocal) resolver(chave string) (string, error) {
	chave = strings.ReplaceAll(strings.TrimSpace(chave), "\\", "/")
	limpa := path.Clean("/" + chave)
	if chave == "" || strings.Contains(chave, "..") || limpa == "/" {
		return "", errors.New("chave de objeto inválida")
	}
	destino := filepath.Join(a.raiz, filepath.FromSlash(strings.TrimPrefix(limpa, "/")))
	relativo, err := filepath.Rel(a.raiz, destino)
	if err != nil || relativo == ".." || strings.HasPrefix(relativo, ".."+string(filepath.Separator)) {
		return "", errors.New("chave fora do armazenamento")
	}
	return destino, nil
}

func (a *ArmazenamentoLocal) urlPublica(chave string) string {
	segmentos := strings.Split(strings.Trim(chave, "/"), "/")
	for indice := range segmentos {
		segmentos[indice] = url.PathEscape(segmentos[indice])
	}
	return strings.TrimRight(a.baseURL, "/") + "/" + strings.Join(segmentos, "/")
}
