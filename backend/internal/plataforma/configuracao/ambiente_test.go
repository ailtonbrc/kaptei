package configuracao

import (
	"net/url"
	"strings"
	"testing"
)

func limparConfiguracaoBanco(t *testing.T) {
	t.Helper()
	for _, chave := range []string{
		"DATABASE_URL", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD",
		"DB_DATABASE", "DB_SSLMODE", "ENV",
	} {
		t.Setenv(chave, "")
	}
}

func TestDatabaseURLExplicitaTemPrioridade(t *testing.T) {
	limparConfiguracaoBanco(t)
	t.Setenv("DATABASE_URL", "postgres://principal:segredo@banco:5432/kaptei?sslmode=require")
	t.Setenv("DB_HOST", "legado")

	resultado, err := obterDatabaseURLDoAmbiente()
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if resultado != "postgres://principal:segredo@banco:5432/kaptei?sslmode=require" {
		t.Fatal("DATABASE_URL explícita deve ter prioridade")
	}
}

func TestDatabaseURLLegadaEscapaCredenciais(t *testing.T) {
	limparConfiguracaoBanco(t)
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "usuario@local")
	t.Setenv("DB_PASSWORD", "s:e/n?h#a")
	t.Setenv("DB_DATABASE", "kaptei_teste")
	t.Setenv("ENV", "development")

	resultado, err := obterDatabaseURLDoAmbiente()
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	endereco, err := url.Parse(resultado)
	if err != nil {
		t.Fatalf("URL montada é inválida: %v", err)
	}
	senha, possuiSenha := endereco.User.Password()
	if endereco.User.Username() != "usuario@local" || !possuiSenha || senha != "s:e/n?h#a" {
		t.Fatal("usuário e senha devem sobreviver ao escape da URL")
	}
	if endereco.Host != "localhost:5433" || endereco.Path != "/kaptei_teste" {
		t.Fatalf("destino PostgreSQL inesperado: %s%s", endereco.Host, endereco.Path)
	}
	if endereco.Query().Get("sslmode") != "disable" {
		t.Fatal("desenvolvimento local deve desabilitar TLS somente quando ENV for explícito")
	}
}

func TestDatabaseURLLegadaExigeConjuntoCompleto(t *testing.T) {
	limparConfiguracaoBanco(t)
	t.Setenv("DB_HOST", "localhost")

	_, err := obterDatabaseURLDoAmbiente()
	if err == nil || !strings.Contains(err.Error(), "DB_USER") || !strings.Contains(err.Error(), "DB_PASSWORD") || !strings.Contains(err.Error(), "DB_DATABASE") {
		t.Fatalf("esperava diagnóstico das variáveis ausentes, recebeu: %v", err)
	}
}

func TestDatabaseURLLegadaUsaTLSPorPadraoForaDoDesenvolvimento(t *testing.T) {
	limparConfiguracaoBanco(t)
	t.Setenv("DB_HOST", "banco.interno")
	t.Setenv("DB_USER", "kaptei")
	t.Setenv("DB_PASSWORD", "segredo")
	t.Setenv("DB_DATABASE", "kaptei")
	t.Setenv("ENV", "production")

	resultado, err := obterDatabaseURLDoAmbiente()
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	endereco, _ := url.Parse(resultado)
	if endereco.Query().Get("sslmode") != "require" {
		t.Fatal("ambientes não locais devem exigir TLS por padrão")
	}
}
