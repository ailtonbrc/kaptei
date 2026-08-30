package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestFiltroPaginacaoAplicaLimites(t *testing.T) {
	requisicao := httptest.NewRequest("GET", "/?pagina=0&limite=999&busca=%20Maria%20&status=NOVO", nil)
	filtro := filtroPaginacao(requisicao)
	if filtro.Pagina != 1 || filtro.Limite != 100 || filtro.Busca != "Maria" || filtro.Status != "NOVO" {
		t.Fatalf("filtro inesperado: %#v", filtro)
	}
}

func TestFiltroPaginacaoUsaPadroesParaValoresInvalidos(t *testing.T) {
	requisicao := httptest.NewRequest("GET", "/?pagina=abc&limite=-1", nil)
	filtro := filtroPaginacao(requisicao)
	if filtro.Pagina != 1 || filtro.Limite != 50 {
		t.Fatalf("padrões inesperados: %#v", filtro)
	}
}
