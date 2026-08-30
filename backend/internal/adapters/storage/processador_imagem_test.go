package storage

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestProcessadorImagemNormalizaERedimensiona(t *testing.T) {
	original := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			original.Set(x, y, color.RGBA{R: 20, G: 100, B: 180, A: 255})
		}
	}
	var entrada bytes.Buffer
	if err := png.Encode(&entrada, original); err != nil {
		t.Fatalf("gerar imagem de teste: %v", err)
	}
	processador := NewProcessadorImagem(1024*1024, 1_000_000, 80, 20, 80)
	resultado, err := processador.Processar(entrada.Bytes())
	if err != nil {
		t.Fatalf("processar imagem: %v", err)
	}
	if resultado.TipoConteudo != "image/jpeg" || resultado.Extensao != ".jpg" {
		t.Fatalf("formato normalizado inesperado: %+v", resultado)
	}
	if resultado.Largura != 80 || resultado.Altura != 40 {
		t.Fatalf("dimensões principais = %dx%d", resultado.Largura, resultado.Altura)
	}
	miniatura, _, err := image.DecodeConfig(bytes.NewReader(resultado.Thumbnail))
	if err != nil || miniatura.Width != 20 || miniatura.Height != 10 {
		t.Fatalf("miniatura inválida: %+v, erro=%v", miniatura, err)
	}
}

func TestOrientacaoEXIFRotacionaImagem(t *testing.T) {
	tiff := make([]byte, 26)
	copy(tiff[:2], "II")
	binary.LittleEndian.PutUint16(tiff[2:4], 42)
	binary.LittleEndian.PutUint32(tiff[4:8], 8)
	binary.LittleEndian.PutUint16(tiff[8:10], 1)
	binary.LittleEndian.PutUint16(tiff[10:12], 0x0112)
	binary.LittleEndian.PutUint16(tiff[12:14], 3)
	binary.LittleEndian.PutUint32(tiff[14:18], 1)
	binary.LittleEndian.PutUint16(tiff[18:20], 6)
	if orientacao := orientacaoTIFF(tiff); orientacao != 6 {
		t.Fatalf("orientação lida = %d, esperado 6", orientacao)
	}
	original := image.NewRGBA(image.Rect(0, 0, 2, 1))
	original.Set(0, 0, color.RGBA{R: 255, A: 255})
	original.Set(1, 0, color.RGBA{B: 255, A: 255})
	rotacionada := aplicarOrientacao(original, 6)
	if rotacionada.Bounds().Dx() != 1 || rotacionada.Bounds().Dy() != 2 {
		t.Fatalf("dimensões após orientação = %v", rotacionada.Bounds())
	}
	if vermelho := color.RGBAModel.Convert(rotacionada.At(0, 0)).(color.RGBA); vermelho.R != 255 {
		t.Fatalf("primeiro pixel não foi preservado: %+v", vermelho)
	}
}

func TestProcessadorImagemRejeitaConteudoArbitrario(t *testing.T) {
	processador := NewProcessadorImagem(1024, 1_000_000, 800, 200, 80)
	if _, err := processador.Processar([]byte("não é uma imagem")); err == nil {
		t.Fatal("era esperado erro para conteúdo arbitrário")
	}
}
