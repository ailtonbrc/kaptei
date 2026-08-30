package services

import (
	"context"
	"errors"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

var slugPublicoValido = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var corHexadecimalValida = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var caracteresNaoNumericos = regexp.MustCompile(`\D`)

type sitePublicoService struct {
	repositorio ports.SitePublicoRepository
	leads       ports.LeadService
}

func NewSitePublicoService(repositorio ports.SitePublicoRepository, leads ports.LeadService) ports.SitePublicoService {
	return &sitePublicoService{repositorio: repositorio, leads: leads}
}

func (s *sitePublicoService) GetPublico(ctx context.Context, slug string) (*domain.SitePublico, error) {
	return s.repositorio.GetBySlug(ctx, normalizarSlug(slug))
}

func (s *sitePublicoService) ListarRotasSitemap(ctx context.Context) ([]domain.RotaSitemap, error) {
	return s.repositorio.ListarRotasSitemap(ctx)
}

func (s *sitePublicoService) GetAdministracao(ctx context.Context, contaID string) (*domain.SitePublico, error) {
	return s.repositorio.GetByContaID(ctx, contaID)
}

func (s *sitePublicoService) Salvar(ctx context.Context, site *domain.SitePublico) error {
	site.Slug = normalizarSlug(site.Slug)
	if site.Publicado && site.Slug == "" {
		return errors.New("defina o endereço público antes de publicar")
	}
	if site.Slug != "" && (len(site.Slug) < 3 || len(site.Slug) > 80 || !slugPublicoValido.MatchString(site.Slug)) {
		return errors.New("endereço público inválido")
	}
	if err := validarConfiguracaoSite(&site.Configuracao); err != nil {
		return err
	}
	if site.Publicado && site.Configuracao.Email == "" {
		return errors.New("informe um e-mail público para atendimento de privacidade antes de publicar")
	}
	return s.repositorio.Salvar(ctx, site)
}

func (s *sitePublicoService) ListarImoveis(ctx context.Context, slug string, filtros domain.FiltrosCatalogoPublico) ([]*domain.ImovelPublico, int, error) {
	site, err := s.GetPublico(ctx, slug)
	if err != nil || site == nil {
		return nil, 0, err
	}
	normalizarPaginacao(&filtros)
	return s.repositorio.ListarImoveis(ctx, site.ContaID, filtros)
}

func (s *sitePublicoService) GetImovel(ctx context.Context, slugSite, slugImovel string) (*domain.ImovelPublico, error) {
	site, err := s.GetPublico(ctx, slugSite)
	if err != nil || site == nil {
		return nil, err
	}
	return s.repositorio.GetImovelBySlug(ctx, site.ContaID, normalizarSlug(slugImovel))
}

func (s *sitePublicoService) CapturarLead(ctx context.Context, slug string, captura domain.CapturaLeadPublico) error {
	site, err := s.GetPublico(ctx, slug)
	if err != nil || site == nil {
		return errors.New("site não encontrado")
	}
	if strings.TrimSpace(captura.Website) != "" {
		// Campo-armadilha: robôs recebem sucesso sem inserir dados no CRM.
		return nil
	}
	var imovelID *string
	if captura.ImovelSlug != nil && strings.TrimSpace(*captura.ImovelSlug) != "" {
		imovel, err := s.repositorio.GetImovelBySlug(ctx, site.ContaID, normalizarSlug(*captura.ImovelSlug))
		if err != nil {
			return err
		}
		if imovel == nil {
			return errors.New("imóvel não encontrado")
		}
		imovelID = &imovel.ID
	}
	return s.leads.CaptarSite(ctx, site.ContaID, captura, imovelID)
}

func normalizarSlug(slug string) string { return strings.ToLower(strings.TrimSpace(slug)) }

func normalizarPaginacao(filtros *domain.FiltrosCatalogoPublico) {
	if filtros.Pagina < 1 {
		filtros.Pagina = 1
	}
	if filtros.Limite < 1 {
		filtros.Limite = 12
	}
	if filtros.Limite > 48 {
		filtros.Limite = 48
	}
	if filtros.Pagina > 100_000 {
		filtros.Pagina = 100_000
	}
}

func validarConfiguracaoSite(configuracao *domain.ConfiguracaoSitePublico) error {
	configuracao.LogoURL = strings.TrimSpace(configuracao.LogoURL)
	configuracao.CorPrimaria = strings.TrimSpace(configuracao.CorPrimaria)
	configuracao.Titulo = strings.TrimSpace(configuracao.Titulo)
	configuracao.Subtitulo = strings.TrimSpace(configuracao.Subtitulo)
	configuracao.Descricao = strings.TrimSpace(configuracao.Descricao)
	configuracao.Telefone = strings.TrimSpace(configuracao.Telefone)
	configuracao.WhatsApp = strings.TrimSpace(configuracao.WhatsApp)
	configuracao.Email = strings.ToLower(strings.TrimSpace(configuracao.Email))
	configuracao.Cidade = strings.TrimSpace(configuracao.Cidade)
	configuracao.CRECI = strings.TrimSpace(configuracao.CRECI)

	limites := []struct {
		rotulo string
		valor  string
		limite int
	}{
		{"título", configuracao.Titulo, 120}, {"subtítulo", configuracao.Subtitulo, 220},
		{"apresentação", configuracao.Descricao, 1000}, {"telefone", configuracao.Telefone, 30},
		{"WhatsApp", configuracao.WhatsApp, 30}, {"e-mail", configuracao.Email, 254},
		{"cidade", configuracao.Cidade, 120}, {"CRECI", configuracao.CRECI, 40},
	}
	for _, campo := range limites {
		if utf8.RuneCountInString(campo.valor) > campo.limite {
			return errors.New(campo.rotulo + " excede o limite permitido")
		}
	}
	if configuracao.CorPrimaria != "" && !corHexadecimalValida.MatchString(configuracao.CorPrimaria) {
		return errors.New("cor primária deve usar o formato hexadecimal #RRGGBB")
	}
	if configuracao.LogoURL != "" {
		if len(configuracao.LogoURL) > 2048 || !urlHTTPValida(configuracao.LogoURL) {
			return errors.New("URL do logotipo inválida")
		}
	}
	if configuracao.Email != "" {
		endereco, err := mail.ParseAddress(configuracao.Email)
		if err != nil || !strings.EqualFold(endereco.Address, configuracao.Email) {
			return errors.New("e-mail público inválido")
		}
	}
	for _, contato := range []struct {
		nome  string
		valor string
	}{{"telefone", configuracao.Telefone}, {"WhatsApp", configuracao.WhatsApp}} {
		if contato.valor == "" {
			continue
		}
		digitos := caracteresNaoNumericos.ReplaceAllString(contato.valor, "")
		if len(digitos) < 8 || len(digitos) > 15 {
			return errors.New(contato.nome + " inválido")
		}
	}
	return nil
}

func urlHTTPValida(valor string) bool {
	endereco, err := url.ParseRequestURI(valor)
	return err == nil && endereco.IsAbs() && endereco.Host != "" && endereco.User == nil &&
		(endereco.Scheme == "http" || endereco.Scheme == "https")
}
