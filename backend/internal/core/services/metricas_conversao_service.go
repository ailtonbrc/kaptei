package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

var (
	ErrEventoConversaoInvalido    = errors.New("evento de conversão inválido")
	ErrSiteConversaoNaoEncontrado = errors.New("site não encontrado")
)

type metricasConversaoService struct {
	sites       ports.SitePublicoRepository
	repositorio ports.MetricasConversaoRepository
}

func NewMetricasConversaoService(sites ports.SitePublicoRepository, repositorio ports.MetricasConversaoRepository) ports.MetricasConversaoService {
	return &metricasConversaoService{sites: sites, repositorio: repositorio}
}

func (s *metricasConversaoService) Registrar(ctx context.Context, slug string, evento domain.EventoConversaoPublico) error {
	if !formatoUUID.MatchString(strings.TrimSpace(evento.ChaveEvento)) || !formatoUUID.MatchString(strings.TrimSpace(evento.SessaoID)) {
		return ErrEventoConversaoInvalido
	}
	tiposPermitidos := map[domain.TipoEventoConversao]bool{
		domain.EventoSiteVisualizado: true, domain.EventoImovelVisualizado: true,
		domain.EventoFormularioIniciado: true, domain.EventoLeadEnviado: true,
		domain.EventoWhatsAppClicado: true, domain.EventoTelefoneClicado: true,
	}
	if !tiposPermitidos[evento.Tipo] {
		return ErrEventoConversaoInvalido
	}

	site, err := s.sites.GetBySlug(ctx, normalizarSlug(slug))
	if err != nil {
		return err
	}
	if site == nil {
		return ErrSiteConversaoNaoEncontrado
	}

	evento.ChaveEvento = strings.ToLower(strings.TrimSpace(evento.ChaveEvento))
	evento.SessaoID = strings.ToLower(strings.TrimSpace(evento.SessaoID))
	if !normalizarAtribuicao(&evento) {
		return ErrEventoConversaoInvalido
	}

	registro := &domain.EventoConversao{EventoConversaoPublico: evento, ContaID: site.ContaID}
	if evento.ImovelSlug != nil {
		imovel, erroImovel := s.sites.GetImovelBySlug(ctx, site.ContaID, normalizarSlug(*evento.ImovelSlug))
		if erroImovel != nil {
			return erroImovel
		}
		if imovel == nil {
			return ErrEventoConversaoInvalido
		}
		registro.ImovelID = &imovel.ID
	}
	if evento.Tipo == domain.EventoImovelVisualizado && registro.ImovelID == nil {
		return ErrEventoConversaoInvalido
	}
	return s.repositorio.Registrar(ctx, registro)
}

func normalizarAtribuicao(evento *domain.EventoConversaoPublico) bool {
	campos := []struct {
		valor  **string
		limite int
	}{
		{&evento.ImovelSlug, 180}, {&evento.UTMSource, 120},
		{&evento.UTMMedium, 120}, {&evento.UTMCampaign, 180},
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
		if utf8.RuneCountInString(texto) > campo.limite {
			return false
		}
		*campo.valor = &texto
	}
	return true
}

type ProcessadorExpurgoMetricasConversao struct {
	repositorio ports.MetricasConversaoRepository
	intervalo   time.Duration
}

func NewProcessadorExpurgoMetricasConversao(repositorio ports.MetricasConversaoRepository, intervalo time.Duration) *ProcessadorExpurgoMetricasConversao {
	return &ProcessadorExpurgoMetricasConversao{repositorio: repositorio, intervalo: intervalo}
}

func (p *ProcessadorExpurgoMetricasConversao) Executar(ctx context.Context) {
	p.expurgar(ctx)
	ticker := time.NewTicker(p.intervalo)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.expurgar(ctx)
		}
	}
}

func (p *ProcessadorExpurgoMetricasConversao) expurgar(ctx context.Context) {
	total, err := p.repositorio.ExpurgarExpirados(ctx)
	if err != nil {
		slog.Error("falha ao expurgar métricas de conversão", "erro", err)
		return
	}
	if total > 0 {
		slog.Info("métricas de conversão expiradas removidas", "total", total)
	}
}
