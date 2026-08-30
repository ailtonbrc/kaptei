package main

import "testing"

func TestPartesRotaDominio(t *testing.T) {
	casos := []struct {
		caminho, imovel string
		privacidade     bool
	}{
		{"/", "", false}, {"/privacidade", "", true}, {"/imoveis/casa-centro", "casa-centro", false},
	}
	for _, caso := range casos {
		rota := partesRotaDominio(caso.caminho)
		if rota == nil || rota.imovel != caso.imovel || rota.privacidade != caso.privacidade {
			t.Fatalf("rota inesperada para %q: %#v", caso.caminho, rota)
		}
	}
	if partesRotaDominio("/app") != nil {
		t.Fatal("rota administrativa não pode ser tratada como site por domínio")
	}
}

func TestHostnameRequisicao(t *testing.T) {
	host, ok := hostnameRequisicao("Imoveis.Exemplo.com.br:443")
	if !ok || host != "imoveis.exemplo.com.br" {
		t.Fatalf("host inesperado: %q", host)
	}
	for _, invalido := range []string{"", "exemplo.com/rota", "exemplo.com@malicioso.com", "exemplo.com:porta"} {
		if _, ok := hostnameRequisicao(invalido); ok {
			t.Fatalf("host deveria ser rejeitado: %q", invalido)
		}
	}
}
