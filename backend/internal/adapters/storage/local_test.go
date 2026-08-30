package storage

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestArmazenamentoLocalSalvaEExcluiObjeto(t *testing.T) {
	raiz := t.TempDir()
	objetos, err := NewArmazenamentoLocal(raiz, "http://localhost:8080/arquivos")
	if err != nil {
		t.Fatalf("criar armazenamento: %v", err)
	}
	url, err := objetos.Salvar(context.Background(), "conta/imoveis/foto.jpg", []byte("imagem"), "image/jpeg")
	if err != nil {
		t.Fatalf("salvar objeto: %v", err)
	}
	if url != "http://localhost:8080/arquivos/conta/imoveis/foto.jpg" {
		t.Fatalf("URL inesperada: %s", url)
	}
	if _, err := os.Stat(raiz + string(os.PathSeparator) + "conta" + string(os.PathSeparator) + "imoveis" + string(os.PathSeparator) + "foto.jpg"); err != nil {
		t.Fatalf("objeto não foi persistido: %v", err)
	}
	if err := objetos.Excluir(context.Background(), "conta/imoveis/foto.jpg"); err != nil {
		t.Fatalf("excluir objeto: %v", err)
	}
}

func TestArmazenamentoLocalBloqueiaTravessiaDeDiretorio(t *testing.T) {
	objetos, err := NewArmazenamentoLocal(t.TempDir(), "http://localhost:8080/arquivos")
	if err != nil {
		t.Fatalf("criar armazenamento: %v", err)
	}
	_, err = objetos.Salvar(context.Background(), "../segredo.txt", []byte("segredo"), "text/plain")
	if err == nil || !strings.Contains(err.Error(), "inválida") {
		t.Fatalf("travessia deveria ser bloqueada: %v", err)
	}
}
