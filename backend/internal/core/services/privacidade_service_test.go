package services

import (
	"strings"
	"testing"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
)

func TestNormalizarSolicitacaoTitular(t *testing.T) {
	nova := domain.NovaSolicitacaoTitular{
		Tipo: domain.TipoAcesso, Nome: "  Maria da Silva  ", Email: "  MARIA@EXEMPLO.COM ",
		Telefone: "+55 (65) 99999-0000", Detalhes: "  Quero uma cópia.  ",
	}
	if err := normalizarSolicitacaoTitular(&nova); err != nil {
		t.Fatalf("normalizar solicitação válida: %v", err)
	}
	if nova.Nome != "Maria da Silva" || nova.Email != "maria@exemplo.com" || nova.Telefone != "5565999990000" || nova.Detalhes != "Quero uma cópia." {
		t.Fatalf("normalização inesperada: %#v", nova)
	}
}

func TestNormalizarSolicitacaoTitularExigeContato(t *testing.T) {
	nova := domain.NovaSolicitacaoTitular{Tipo: domain.TipoExclusao, Nome: "Maria Silva"}
	if err := normalizarSolicitacaoTitular(&nova); err == nil {
		t.Fatal("esperava erro quando nenhum contato é informado")
	}
}

func TestHashOpcionalNaoExpoeValorOriginal(t *testing.T) {
	valor := "maria@exemplo.com"
	hash := hashOpcional(valor)
	if hash == nil || len(*hash) != 64 || strings.Contains(*hash, valor) {
		t.Fatalf("hash inválido: %v", hash)
	}
	if repetido := hashOpcional(valor); repetido == nil || *repetido != *hash {
		t.Fatal("hash de busca precisa ser determinístico")
	}
}

func TestGerarProtocoloTitular(t *testing.T) {
	agora := time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)
	primeiro, err := gerarProtocoloTitular(agora)
	if err != nil {
		t.Fatalf("gerar protocolo: %v", err)
	}
	segundo, err := gerarProtocoloTitular(agora)
	if err != nil {
		t.Fatalf("gerar segundo protocolo: %v", err)
	}
	if !strings.HasPrefix(primeiro, "KPT-20260806-") || primeiro == segundo {
		t.Fatalf("protocolos inesperados: %q e %q", primeiro, segundo)
	}
}
