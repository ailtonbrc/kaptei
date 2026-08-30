package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"
	"net/url"
	"regexp"
	"strings"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type imovelService struct {
	repo              ports.ImovelRepository
	processadorImagem ports.ProcessadorImagem
	objetos           ports.ArmazenamentoObjetos
	preparadorObjeto  ports.PreparadorObjetoOutbox
}

func NewImovelService(
	repo ports.ImovelRepository,
	processadorImagem ports.ProcessadorImagem,
	objetos ports.ArmazenamentoObjetos,
	preparadorObjeto ports.PreparadorObjetoOutbox,
) ports.ImovelService {
	return &imovelService{
		repo: repo, processadorImagem: processadorImagem,
		objetos: objetos, preparadorObjeto: preparadorObjeto,
	}
}

func (s *imovelService) Create(ctx context.Context, imovel *domain.Imovel) error {
	if imovel.ContaID == "" {
		return errors.New("conta_id é obrigatório")
	}
	if imovel.UsuarioID == "" {
		return errors.New("usuario_id é obrigatório")
	}
	if imovel.Status == "" {
		imovel.Status = "Ativo"
	}
	if err := validarImovel(imovel); err != nil {
		return err
	}

	return s.repo.Create(ctx, imovel)
}

func (s *imovelService) GetByID(ctx context.Context, id, contaID string) (*domain.Imovel, error) {
	if id == "" || contaID == "" {
		return nil, errors.New("id e conta_id são obrigatórios")
	}
	return s.repo.GetByID(ctx, id, contaID)
}

func (s *imovelService) List(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.Imovel], error) {
	if contaID == "" {
		return nil, errors.New("conta_id é obrigatório")
	}
	filtro = normalizarFiltroPaginacao(filtro)
	return s.repo.ListByContaID(ctx, contaID, filtro)
}

func (s *imovelService) Update(ctx context.Context, imovel *domain.Imovel) error {
	if imovel.ID == "" || imovel.ContaID == "" {
		return errors.New("id e conta_id são obrigatórios para atualização")
	}
	if err := validarImovel(imovel); err != nil {
		return err
	}
	return s.repo.Update(ctx, imovel)
}

func (s *imovelService) Delete(ctx context.Context, id, contaID string) error {
	if id == "" || contaID == "" {
		return errors.New("id e conta_id são obrigatórios para exclusão")
	}
	imovel, err := s.repo.GetByID(ctx, id, contaID)
	if err != nil {
		return err
	}
	if imovel == nil {
		return errors.New("imóvel não encontrado")
	}
	eventos, err := s.prepararExclusoesFotos(contaID, imovel.Fotos)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id, contaID, eventos)
}

func (s *imovelService) AddFoto(ctx context.Context, imovelID, enderecoURL string, isCapa bool) (*domain.ImovelFoto, error) {
	if imovelID == "" || enderecoURL == "" {
		return nil, errors.New("imovel_id e url são obrigatórios")
	}
	enderecoURL = strings.TrimSpace(enderecoURL)
	urlFoto, err := url.ParseRequestURI(enderecoURL)
	if err != nil || len(enderecoURL) > 2048 || urlFoto.Scheme != "https" || urlFoto.Host == "" || urlFoto.User != nil {
		return nil, errors.New("URL da foto inválida")
	}

	foto := &domain.ImovelFoto{
		ImovelID: imovelID,
		URL:      enderecoURL,
		IsCapa:   isCapa,
	}

	err = s.repo.AddFoto(ctx, foto)
	if err != nil {
		return nil, err
	}
	return foto, nil
}

func (s *imovelService) UploadFoto(ctx context.Context, imovelID, contaID string, conteudo []byte, isCapa bool) (*domain.ImovelFoto, error) {
	if imovelID == "" || contaID == "" {
		return nil, errors.New("imóvel e conta são obrigatórios")
	}
	imovel, err := s.repo.GetByID(ctx, imovelID, contaID)
	if err != nil {
		return nil, err
	}
	if imovel == nil {
		return nil, errors.New("imóvel não encontrado")
	}
	imagem, err := s.processadorImagem.Processar(conteudo)
	if err != nil {
		return nil, err
	}
	identificador, err := gerarIdentificadorObjeto()
	if err != nil {
		return nil, errors.New("não foi possível gerar o identificador da imagem")
	}
	prefixo := contaID + "/imoveis/" + imovelID + "/" + identificador
	chavePrincipal := prefixo + "-principal" + imagem.Extensao
	chaveThumbnail := prefixo + "-thumbnail" + imagem.Extensao
	urlPrincipal, err := s.objetos.Salvar(ctx, chavePrincipal, imagem.Principal, imagem.TipoConteudo)
	if err != nil {
		return nil, err
	}
	urlThumbnail, err := s.objetos.Salvar(ctx, chaveThumbnail, imagem.Thumbnail, imagem.TipoConteudo)
	if err != nil {
		_ = s.objetos.Excluir(context.Background(), chavePrincipal)
		return nil, err
	}
	provedor := s.objetos.Nome()
	tipoConteudo := imagem.TipoConteudo
	tamanho := int64(len(imagem.Principal))
	largura, altura, hash := imagem.Largura, imagem.Altura, imagem.HashSHA256
	foto := &domain.ImovelFoto{
		ImovelID: imovelID, URL: urlPrincipal, URLThumbnail: &urlThumbnail,
		ChaveObjeto: &chavePrincipal, ChaveThumbnail: &chaveThumbnail, ProvedorStorage: &provedor,
		TipoConteudo: &tipoConteudo, TamanhoBytes: &tamanho, Largura: &largura, Altura: &altura,
		HashSHA256: &hash, IsCapa: isCapa,
	}
	if err := s.repo.AddFoto(ctx, foto); err != nil {
		_ = s.objetos.Excluir(context.Background(), chavePrincipal)
		_ = s.objetos.Excluir(context.Background(), chaveThumbnail)
		return nil, err
	}
	return foto, nil
}

var slugImovelValido = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validarImovel(imovel *domain.Imovel) error {
	imovel.Titulo = strings.TrimSpace(imovel.Titulo)
	imovel.Tipo = strings.TrimSpace(imovel.Tipo)
	imovel.Finalidade = strings.TrimSpace(imovel.Finalidade)
	imovel.Status = strings.TrimSpace(imovel.Status)
	if imovel.Titulo == "" || len([]rune(imovel.Titulo)) > 255 {
		return errors.New("título do imóvel excede 255 caracteres")
	}
	tiposValidos := map[string]bool{"Casa": true, "Apartamento": true, "Terreno": true, "Comercial": true, "Galpão": true, "Rural": true}
	if !tiposValidos[imovel.Tipo] {
		return errors.New("tipo do imóvel inválido")
	}
	finalidadesValidas := map[string]bool{"Venda": true, "Locação": true, "Venda e Locação": true}
	if !finalidadesValidas[imovel.Finalidade] {
		return errors.New("finalidade do imóvel inválida")
	}
	statusValidos := map[string]bool{"Ativo": true, "Inativo": true, "Vendido": true, "Alugado": true}
	if !statusValidos[imovel.Status] {
		return errors.New("status do imóvel inválido")
	}
	for _, numero := range []*float64{imovel.ValorVenda, imovel.ValorLocacao, imovel.ValorCondominio, imovel.ValorIPTU} {
		if numero != nil && (math.IsNaN(*numero) || math.IsInf(*numero, 0) || *numero < 0 || *numero > 999_999_999_999) {
			return errors.New("valor monetário do imóvel inválido")
		}
	}
	for _, area := range []*float64{imovel.AreaTotal, imovel.AreaUtil} {
		if area != nil && (math.IsNaN(*area) || math.IsInf(*area, 0) || *area < 0 || *area > 99_999_999) {
			return errors.New("área do imóvel inválida")
		}
	}
	if imovel.Quartos < 0 || imovel.Quartos > 100 || imovel.Suites < 0 || imovel.Suites > 100 ||
		imovel.Banheiros < 0 || imovel.Banheiros > 100 || imovel.Vagas < 0 || imovel.Vagas > 100 {
		return errors.New("quantidade de cômodos ou vagas inválida")
	}
	campos := []struct {
		valor  **string
		limite int
		nome   string
	}{
		{&imovel.CEP, 20, "CEP"}, {&imovel.Logradouro, 255, "logradouro"}, {&imovel.Numero, 50, "número"},
		{&imovel.Complemento, 150, "complemento"}, {&imovel.Bairro, 150, "bairro"}, {&imovel.Cidade, 150, "cidade"},
		{&imovel.Estado, 2, "estado"}, {&imovel.Descricao, 10_000, "descrição"},
	}
	for _, campo := range campos {
		if *campo.valor == nil {
			continue
		}
		texto := strings.TrimSpace(**campo.valor)
		if texto == "" {
			*campo.valor = nil
			continue
		}
		if len([]rune(texto)) > campo.limite {
			return errors.New(campo.nome + " excede o limite permitido")
		}
		*campo.valor = &texto
	}
	if imovel.SlugPublico != nil {
		slug := strings.ToLower(strings.TrimSpace(*imovel.SlugPublico))
		if slug == "" {
			imovel.SlugPublico = nil
		} else {
			if len(slug) > 180 || !slugImovelValido.MatchString(slug) {
				return errors.New("endereço público do imóvel é inválido")
			}
			imovel.SlugPublico = &slug
		}
	}
	if imovel.Publicado {
		if imovel.Status != "Ativo" || imovel.SlugPublico == nil {
			return errors.New("para publicar, o imóvel deve estar ativo e possuir endereço público")
		}
	}
	if imovel.TituloSEO != nil && len([]rune(*imovel.TituloSEO)) > 180 {
		return errors.New("título SEO excede 180 caracteres")
	}
	if imovel.DescricaoSEO != nil && len([]rune(*imovel.DescricaoSEO)) > 320 {
		return errors.New("descrição SEO excede 320 caracteres")
	}
	return nil
}

func (s *imovelService) DeleteFoto(ctx context.Context, fotoID, imovelID, contaID string) error {
	if fotoID == "" || imovelID == "" || contaID == "" {
		return errors.New("foto, imóvel e conta são obrigatórios")
	}
	imovel, err := s.repo.GetByID(ctx, imovelID, contaID)
	if err != nil || imovel == nil {
		return errors.New("imóvel não encontrado")
	}
	foto, err := s.repo.GetFoto(ctx, fotoID, imovelID)
	if err != nil {
		return err
	}
	if foto == nil {
		return errors.New("foto não encontrada")
	}
	eventos, err := s.prepararExclusoesFotos(contaID, []domain.ImovelFoto{*foto})
	if err != nil {
		return err
	}
	return s.repo.DeleteFoto(ctx, fotoID, imovelID, eventos)
}

func (s *imovelService) prepararExclusoesFotos(contaID string, fotos []domain.ImovelFoto) ([]*domain.EventoOutbox, error) {
	eventos := make([]*domain.EventoOutbox, 0, len(fotos)*2)
	for _, foto := range fotos {
		if foto.ProvedorStorage == nil {
			continue
		}
		for sufixo, chave := range map[string]*string{"principal": foto.ChaveObjeto, "thumbnail": foto.ChaveThumbnail} {
			if chave == nil || strings.TrimSpace(*chave) == "" {
				continue
			}
			evento, err := s.preparadorObjeto.PrepararExclusaoObjeto(
				contaID, "foto:"+foto.ID+":"+sufixo, *foto.ProvedorStorage, *chave,
			)
			if err != nil {
				return nil, err
			}
			eventos = append(eventos, evento)
		}
	}
	return eventos, nil
}

func gerarIdentificadorObjeto() (string, error) {
	bytesAleatorios := make([]byte, 16)
	if _, err := rand.Read(bytesAleatorios); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytesAleatorios), nil
}
