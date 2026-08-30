package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	stddraw "image/draw"
	"image/jpeg"
	_ "image/png"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

type processadorImagem struct {
	maximoBytes     int
	maximoPixels    int64
	maximoPrincipal int
	maximoThumbnail int
	qualidade       int
}

func NewProcessadorImagem(maximoBytes int, maximoPixels int64, maximoPrincipal, maximoThumbnail, qualidade int) ports.ProcessadorImagem {
	return &processadorImagem{
		maximoBytes: maximoBytes, maximoPixels: maximoPixels,
		maximoPrincipal: maximoPrincipal, maximoThumbnail: maximoThumbnail, qualidade: qualidade,
	}
}

func (p *processadorImagem) Processar(conteudo []byte) (*domain.ImagemProcessada, error) {
	if len(conteudo) == 0 || len(conteudo) > p.maximoBytes {
		return nil, errors.New("imagem vazia ou acima do limite permitido")
	}
	configuracao, formato, err := image.DecodeConfig(bytes.NewReader(conteudo))
	if err != nil || (formato != "jpeg" && formato != "png" && formato != "webp") {
		return nil, errors.New("envie uma imagem JPEG, PNG ou WebP válida")
	}
	if configuracao.Width <= 0 || configuracao.Height <= 0 || int64(configuracao.Width)*int64(configuracao.Height) > p.maximoPixels {
		return nil, errors.New("dimensões da imagem excedem o limite permitido")
	}
	original, formatoDecodificado, err := image.Decode(bytes.NewReader(conteudo))
	if err != nil || formatoDecodificado != formato {
		return nil, errors.New("não foi possível decodificar a imagem")
	}
	if formato == "jpeg" {
		original = aplicarOrientacao(original, lerOrientacaoEXIF(conteudo))
	}
	principal := redimensionar(original, p.maximoPrincipal)
	thumbnail := redimensionar(original, p.maximoThumbnail)
	principalJPEG, err := codificarJPEG(principal, p.qualidade)
	if err != nil {
		return nil, fmt.Errorf("gerar imagem principal: %w", err)
	}
	thumbnailJPEG, err := codificarJPEG(thumbnail, p.qualidade)
	if err != nil {
		return nil, fmt.Errorf("gerar miniatura: %w", err)
	}
	hash := sha256.Sum256(principalJPEG)
	return &domain.ImagemProcessada{
		Principal: principalJPEG, Thumbnail: thumbnailJPEG,
		TipoConteudo: "image/jpeg", Extensao: ".jpg",
		Largura: principal.Bounds().Dx(), Altura: principal.Bounds().Dy(),
		HashSHA256: fmt.Sprintf("%x", hash),
	}, nil
}

func lerOrientacaoEXIF(conteudo []byte) int {
	if len(conteudo) < 4 || conteudo[0] != 0xff || conteudo[1] != 0xd8 {
		return 1
	}
	for posicao := 2; posicao+4 <= len(conteudo); {
		if conteudo[posicao] != 0xff {
			break
		}
		marcador := conteudo[posicao+1]
		posicao += 2
		if marcador == 0xd9 || marcador == 0xda {
			break
		}
		if posicao+2 > len(conteudo) {
			break
		}
		tamanho := int(binary.BigEndian.Uint16(conteudo[posicao : posicao+2]))
		if tamanho < 2 || posicao+tamanho > len(conteudo) {
			break
		}
		segmento := conteudo[posicao+2 : posicao+tamanho]
		if marcador == 0xe1 && len(segmento) > 14 && bytes.Equal(segmento[:6], []byte("Exif\x00\x00")) {
			if orientacao := orientacaoTIFF(segmento[6:]); orientacao >= 1 && orientacao <= 8 {
				return orientacao
			}
		}
		posicao += tamanho
	}
	return 1
}

func orientacaoTIFF(tiff []byte) int {
	if len(tiff) < 8 {
		return 1
	}
	var ordem binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		ordem = binary.LittleEndian
	case "MM":
		ordem = binary.BigEndian
	default:
		return 1
	}
	if ordem.Uint16(tiff[2:4]) != 42 {
		return 1
	}
	offset := int(ordem.Uint32(tiff[4:8]))
	if offset < 0 || offset+2 > len(tiff) {
		return 1
	}
	quantidade := int(ordem.Uint16(tiff[offset : offset+2]))
	for indice := 0; indice < quantidade; indice++ {
		inicio := offset + 2 + indice*12
		if inicio+12 > len(tiff) {
			break
		}
		entrada := tiff[inicio : inicio+12]
		if ordem.Uint16(entrada[:2]) == 0x0112 && ordem.Uint16(entrada[2:4]) == 3 && ordem.Uint32(entrada[4:8]) == 1 {
			return int(ordem.Uint16(entrada[8:10]))
		}
	}
	return 1
}

func aplicarOrientacao(origem image.Image, orientacao int) image.Image {
	if orientacao <= 1 || orientacao > 8 {
		return origem
	}
	limites := origem.Bounds()
	largura, altura := limites.Dx(), limites.Dy()
	larguraDestino, alturaDestino := largura, altura
	if orientacao >= 5 {
		larguraDestino, alturaDestino = altura, largura
	}
	destino := image.NewRGBA(image.Rect(0, 0, larguraDestino, alturaDestino))
	for y := 0; y < altura; y++ {
		for x := 0; x < largura; x++ {
			destinoX, destinoY := coordenadaOrientada(x, y, largura, altura, orientacao)
			destino.Set(destinoX, destinoY, origem.At(limites.Min.X+x, limites.Min.Y+y))
		}
	}
	return destino
}

func coordenadaOrientada(x, y, largura, altura, orientacao int) (int, int) {
	switch orientacao {
	case 2:
		return largura - 1 - x, y
	case 3:
		return largura - 1 - x, altura - 1 - y
	case 4:
		return x, altura - 1 - y
	case 5:
		return y, x
	case 6:
		return altura - 1 - y, x
	case 7:
		return altura - 1 - y, largura - 1 - x
	case 8:
		return y, largura - 1 - x
	default:
		return x, y
	}
}

func redimensionar(origem image.Image, limite int) image.Image {
	largura, altura := origem.Bounds().Dx(), origem.Bounds().Dy()
	if largura <= limite && altura <= limite {
		return origem
	}
	novaLargura, novaAltura := largura, altura
	if largura >= altura {
		novaLargura = limite
		novaAltura = max(1, int(float64(altura)*float64(limite)/float64(largura)))
	} else {
		novaAltura = limite
		novaLargura = max(1, int(float64(largura)*float64(limite)/float64(altura)))
	}
	destino := image.NewRGBA(image.Rect(0, 0, novaLargura, novaAltura))
	xdraw.CatmullRom.Scale(destino, destino.Bounds(), origem, origem.Bounds(), stddraw.Src, nil)
	return destino
}

func codificarJPEG(origem image.Image, qualidade int) ([]byte, error) {
	// JPEG não possui transparência. Um fundo branco evita áreas transparentes pretas.
	destino := image.NewRGBA(origem.Bounds())
	stddraw.Draw(destino, destino.Bounds(), &image.Uniform{C: color.White}, image.Point{}, stddraw.Src)
	stddraw.Draw(destino, destino.Bounds(), origem, origem.Bounds().Min, stddraw.Over)
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, destino, &jpeg.Options{Quality: qualidade}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
