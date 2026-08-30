package email

import "testing"

func TestNormalizarIDMensagem(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nome     string
		entrada  string
		esperado string
	}{
		{nome: "vazio", entrada: "  ", esperado: ""},
		{nome: "com delimitadores", entrada: " <evento@outbox.kaptei> ", esperado: "evento@outbox.kaptei"},
		{nome: "sem delimitadores", entrada: "evento@outbox.kaptei", esperado: "evento@outbox.kaptei"},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if resultado := normalizarIDMensagem(caso.entrada); resultado != caso.esperado {
				t.Fatalf("normalizarIDMensagem() = %q; esperado %q", resultado, caso.esperado)
			}
		})
	}
}
