package vrsync

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
)

func textoTeste(valor string) *string    { return &valor }
func numeroTeste(valor float64) *float64 { return &valor }
func inteiro64Teste(valor int64) *int64  { return &valor }

func dadosFeedValidos() *domain.DadosFeedPortal {
	tipoJPEG := "image/jpeg"
	return &domain.DadosFeedPortal{
		Configuracao: domain.ConfiguracaoPortal{
			Ativa: true, NomeContato: "Imobiliária Exemplo", EmailContato: "contato@exemplo.com.br",
			TelefoneContato: "6530000000", ExibicaoEndereco: "BAIRRO",
		},
		NomeConta: "Imobiliária Exemplo", SiteSlug: "imobiliaria-exemplo", SitePublicado: true,
		Anuncios: []domain.AnuncioPortal{{
			ID: "00000000-0000-4000-8000-000000000001", Slug: "apartamento-centro",
			Titulo: "Apartamento amplo no centro", Tipo: "Apartamento", Finalidade: "Venda", Status: "Ativo",
			ValorVenda: numeroTeste(650000.99), ValorCondominio: numeroTeste(750), ValorIPTU: numeroTeste(1200),
			AreaTotal: numeroTeste(100), AreaUtil: numeroTeste(82), Quartos: 3, Suites: 1, Banheiros: 2, Vagas: 2,
			CEP: textoTeste("78000-000"), Logradouro: textoTeste("Avenida Central"), Numero: textoTeste("100"),
			Bairro: textoTeste("Centro"), Cidade: textoTeste("Cuiabá"), Estado: textoTeste("MT"),
			Descricao:      textoTeste("Apartamento bem localizado, com ambientes amplos, ventilação natural e excelente acesso aos serviços do centro."),
			TipoPublicacao: "STANDARD",
			Fotos:          []domain.ImovelFoto{{URL: "https://cdn.exemplo.com.br/foto.jpg", TipoConteudo: &tipoJPEG, TamanhoBytes: inteiro64Teste(500000), IsCapa: true}},
		}},
	}
}

func TestGeradorVRSyncGeraFeedCompleto(t *testing.T) {
	gerador := NovoGerador()
	conteudo, err := gerador.Gerar(dadosFeedValidos(), "https://app.exemplo.com.br", time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}

	var raiz struct{ XMLName xml.Name }
	if err := xml.Unmarshal(conteudo, &raiz); err != nil {
		t.Fatalf("XML gerado é inválido: %v", err)
	}
	if raiz.XMLName.Local != "ListingDataFeed" {
		t.Fatalf("raiz inesperada: %s", raiz.XMLName.Local)
	}
	texto := string(conteudo)
	for _, esperado := range []string{"<ListingID>00000000-0000-4000-8000-000000000001</ListingID>", "<TransactionType>For Sale</TransactionType>", "<PropertyType>Residential / Apartment</PropertyType>", "<ListPrice currency=\"BRL\">650000</ListPrice>", "<![CDATA["} {
		if !strings.Contains(texto, esperado) {
			t.Errorf("feed não contém %q", esperado)
		}
	}
}

func TestGeradorVRSyncRecusaCargaParcial(t *testing.T) {
	dados := dadosFeedValidos()
	dados.Anuncios = append(dados.Anuncios, domain.AnuncioPortal{ID: "00000000-0000-4000-8000-000000000002", Titulo: "Inválido", Status: "Ativo", TipoPublicacao: "STANDARD"})

	diagnostico := NovoGerador().Validar(dados)
	if diagnostico.Valido {
		t.Fatal("carga com anúncio inválido não pode ser considerada válida")
	}
	if diagnostico.TotalSelecionado != 2 || diagnostico.TotalValido != 1 {
		t.Fatalf("contadores inesperados: %+v", diagnostico)
	}
	if _, err := NovoGerador().Gerar(dados, "https://app.exemplo.com.br", time.Now()); err == nil {
		t.Fatal("geração parcial deveria falhar")
	}
}

func TestGeradorVRSyncExigeMetadadosJPEG(t *testing.T) {
	dados := dadosFeedValidos()
	dados.Anuncios[0].Fotos[0].TipoConteudo = nil
	if NovoGerador().Validar(dados).Valido {
		t.Fatal("foto sem metadado de formato não deve ser presumida como JPEG")
	}
}
