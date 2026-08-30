package whatsapp

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
)

type protetorTeste struct{}

func (protetorTeste) Proteger(valor string) (string, error) {
	return "cifrado:" + base64.StdEncoding.EncodeToString([]byte(valor)), nil
}
func (protetorTeste) Revelar(valor string) (string, error) {
	conteudo, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(valor, "cifrado:"))
	return string(conteudo), err
}

func TestCodecWhatsAppProtegeERestauraMensagem(t *testing.T) {
	codec := NewCodecIntegracao(protetorTeste{}, 8)
	original := domain.MensagemWhatsAppRecebida{IdentificadorExterno: "wamid.1", NumeroTelefoneID: "123",
		NumeroContato: "5565999999999", NomeContato: "Maria", Tipo: "text", Texto: "Tenho interesse", OcorridaEm: time.Now().UTC()}
	evento, err := codec.PrepararMensagem("conta-1", original)
	if err != nil {
		t.Fatalf("preparar mensagem: %v", err)
	}
	if strings.Contains(evento.PayloadProtegido, original.Texto) {
		t.Fatal("texto permaneceu legível no evento")
	}
	restaurada, err := codec.DecodificarMensagem(evento)
	if err != nil {
		t.Fatalf("decodificar mensagem: %v", err)
	}
	if restaurada.Texto != original.Texto || restaurada.NumeroContato != original.NumeroContato {
		t.Fatalf("mensagem restaurada diverge: %+v", restaurada)
	}
}
