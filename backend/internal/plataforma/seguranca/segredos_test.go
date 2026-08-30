package seguranca

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestProtetorSegredosProtegeERevela(t *testing.T) {
	chave := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	protetor, err := NovoProtetorSegredos(chave)
	if err != nil {
		t.Fatalf("criar protetor: %v", err)
	}
	cifrado, err := protetor.Proteger("senha-smtp-secreta")
	if err != nil {
		t.Fatalf("proteger segredo: %v", err)
	}
	if !strings.HasPrefix(cifrado, prefixoSegredo) || strings.Contains(cifrado, "senha-smtp-secreta") {
		t.Fatalf("segredo não foi protegido adequadamente: %q", cifrado)
	}
	aberto, err := protetor.Revelar(cifrado)
	if err != nil || aberto != "senha-smtp-secreta" {
		t.Fatalf("revelar segredo: valor=%q erro=%v", aberto, err)
	}
}

func TestProtetorSegredosRejeitaAlteracao(t *testing.T) {
	chave := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	protetor, _ := NovoProtetorSegredos(chave)
	cifrado, _ := protetor.Proteger("valor")
	ultimo := cifrado[len(cifrado)-1]
	substituto := byte('A')
	if ultimo == substituto {
		substituto = 'B'
	}
	adulterado := cifrado[:len(cifrado)-1] + string(substituto)
	if _, err := protetor.Revelar(adulterado); err == nil {
		t.Fatal("segredo adulterado deveria ser rejeitado")
	}
}

func TestNovoProtetorSegredosExigeChaveDe32Bytes(t *testing.T) {
	if _, err := NovoProtetorSegredos(base64.StdEncoding.EncodeToString([]byte("curta"))); err == nil {
		t.Fatal("chave curta deveria ser rejeitada")
	}
}
