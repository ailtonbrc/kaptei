package email

import (
	"encoding/base64"
	"strings"
	"testing"
)

type protetorTeste struct{}

func (protetorTeste) Proteger(valor string) (string, error) {
	return "teste:" + base64.RawStdEncoding.EncodeToString([]byte(valor)), nil
}

func (protetorTeste) Revelar(valor string) (string, error) {
	aberto, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(valor, "teste:"))
	return string(aberto), err
}

func TestCodecOutboxProtegeEDecodificaEmail(t *testing.T) {
	codec := NewCodecOutbox(protetorTeste{}, 8)
	evento, err := codec.PrepararEmail(nil, "recuperacao:hash", "pessoa@example.com", "Assunto", "<p>token-secreto</p>")
	if err != nil {
		t.Fatalf("preparar e-mail: %v", err)
	}
	if strings.Contains(evento.PayloadProtegido, "token-secreto") {
		t.Fatal("o payload persistido não pode expor o conteúdo do e-mail")
	}
	if evento.MaximoTentativas != 8 {
		t.Fatalf("máximo de tentativas = %d, esperado 8", evento.MaximoTentativas)
	}
	mensagem, err := codec.DecodificarEmail(evento)
	if err != nil {
		t.Fatalf("decodificar e-mail: %v", err)
	}
	if mensagem.Destinatario != "pessoa@example.com" || mensagem.CorpoHTML != "<p>token-secreto</p>" {
		t.Fatalf("mensagem decodificada incorretamente: %+v", mensagem)
	}
	if !strings.HasSuffix(mensagem.IDMensagem, "@outbox.kaptei>") {
		t.Fatalf("Message-ID determinístico inválido: %q", mensagem.IDMensagem)
	}
}

func TestCodecOutboxRejeitaDestinatarioInvalido(t *testing.T) {
	codec := NewCodecOutbox(protetorTeste{}, 3)
	if _, err := codec.PrepararEmail(nil, "chave", "invalido", "Assunto", "corpo"); err == nil {
		t.Fatal("era esperado erro para destinatário inválido")
	}
}
