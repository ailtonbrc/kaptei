package whatsapp

import (
	"strings"
	"testing"

	"github.com/msdev/kaptei/internal/core/domain"
)

func TestCodecOutboxProtegeERestauraEnvio(t *testing.T) {
	codec := NewCodecOutbox(protetorTeste{}, 8)
	original := domain.SolicitacaoEnvioWhatsApp{
		IDMensagem: "mensagem-1", ContaID: "conta-1", ConversaID: "conversa-1",
		NumeroTelefoneID: "123456", Destinatario: "5565999999999", Tipo: "TEXTO", Texto: "Olá pelo WhatsApp",
	}
	evento, conteudoProtegido, err := codec.PrepararEnvio(original)
	if err != nil {
		t.Fatalf("preparar envio: %v", err)
	}
	if strings.Contains(conteudoProtegido, original.Texto) {
		t.Fatal("texto permaneceu legível no evento de saída")
	}
	restaurada, err := codec.DecodificarEnvio(evento)
	if err != nil {
		t.Fatalf("decodificar envio: %v", err)
	}
	if restaurada.Texto != original.Texto || restaurada.Destinatario != original.Destinatario {
		t.Fatalf("envio restaurado diverge: %+v", restaurada)
	}
}

func TestCodecOutboxRejeitaEventoDeOutraConta(t *testing.T) {
	codec := NewCodecOutbox(protetorTeste{}, 8)
	evento, _, err := codec.PrepararEnvio(domain.SolicitacaoEnvioWhatsApp{
		IDMensagem: "mensagem-1", ContaID: "conta-1", Destinatario: "5565999999999", Tipo: "TEXTO", Texto: "Teste",
	})
	if err != nil {
		t.Fatalf("preparar envio: %v", err)
	}
	outraConta := "conta-2"
	evento.ContaID = &outraConta
	if _, err := codec.DecodificarEnvio(evento); err == nil {
		t.Fatal("evento adulterado deveria ser rejeitado")
	}
}
