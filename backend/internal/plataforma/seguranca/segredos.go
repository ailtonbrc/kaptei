package seguranca

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const prefixoSegredo = "enc:v1:"

type ProtetorSegredos struct {
	aead cipher.AEAD
}

func NovoProtetorSegredos(chaveBase64 string) (*ProtetorSegredos, error) {
	chave, err := base64.StdEncoding.DecodeString(strings.TrimSpace(chaveBase64))
	if err != nil || len(chave) != 32 {
		return nil, errors.New("CONFIG_ENCRYPTION_KEY deve ser uma chave aleatória de 32 bytes em Base64")
	}
	bloco, err := aes.NewCipher(chave)
	if err != nil {
		return nil, fmt.Errorf("preparar criptografia de segredos: %w", err)
	}
	aead, err := cipher.NewGCM(bloco)
	if err != nil {
		return nil, fmt.Errorf("preparar proteção autenticada: %w", err)
	}
	return &ProtetorSegredos{aead: aead}, nil
}

func (p *ProtetorSegredos) Proteger(valor string) (string, error) {
	nonce := make([]byte, p.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("gerar nonce para o segredo: %w", err)
	}
	cifrado := p.aead.Seal(nil, nonce, []byte(valor), nil)
	pacote := append(nonce, cifrado...)
	return prefixoSegredo + base64.RawStdEncoding.EncodeToString(pacote), nil
}

func (p *ProtetorSegredos) Revelar(valor string) (string, error) {
	if !strings.HasPrefix(valor, prefixoSegredo) {
		// Compatibilidade temporária: o próximo salvamento converte o legado.
		return valor, nil
	}
	pacote, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(valor, prefixoSegredo))
	if err != nil || len(pacote) <= p.aead.NonceSize() {
		return "", errors.New("segredo persistido está corrompido")
	}
	nonce, cifrado := pacote[:p.aead.NonceSize()], pacote[p.aead.NonceSize():]
	aberto, err := p.aead.Open(nil, nonce, cifrado, nil)
	if err != nil {
		return "", errors.New("não foi possível autenticar o segredo persistido")
	}
	return string(aberto), nil
}
