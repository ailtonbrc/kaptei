package services

import "testing"

func TestLimitarRunasNormalizaEspacosEAcentos(t *testing.T) {
	t.Parallel()

	if resultado := limitarRunas("  imóvel amplo  ", 6); resultado != "imóvel" {
		t.Fatalf("limitarRunas() = %q; esperado %q", resultado, "imóvel")
	}
}

func TestLimitarRunasPreservaValorAbaixoDoLimite(t *testing.T) {
	t.Parallel()

	if resultado := limitarRunas("  ATIVO  ", 40); resultado != "ATIVO" {
		t.Fatalf("limitarRunas() = %q; esperado %q", resultado, "ATIVO")
	}
}
