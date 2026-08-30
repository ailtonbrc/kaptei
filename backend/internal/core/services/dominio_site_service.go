package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

var rotuloDominioValido = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

type dominioSiteService struct {
	repo            ports.DominioSiteRepository
	resolvedor      ports.ResolvedorDNS
	hostsReservados map[string]bool
}

func NewDominioSiteService(repo ports.DominioSiteRepository, resolvedor ports.ResolvedorDNS, urlsReservadas ...string) ports.DominioSiteService {
	reservados := make(map[string]bool)
	for _, valor := range urlsReservadas {
		if endereco, err := url.Parse(valor); err == nil && endereco.Hostname() != "" {
			reservados[strings.ToLower(endereco.Hostname())] = true
		}
	}
	return &dominioSiteService{repo: repo, resolvedor: resolvedor, hostsReservados: reservados}
}

func (s *dominioSiteService) Obter(ctx context.Context, contaID string, papel domain.Role) (*domain.DominioSite, error) {
	if !podeAdministrarDominio(papel) {
		return nil, errors.New("sem permissão para administrar domínio")
	}
	dominio, err := s.repo.ObterPorConta(ctx, contaID)
	if dominio != nil {
		prepararInstrucaoDNS(dominio)
	}
	return dominio, err
}

func (s *dominioSiteService) Configurar(ctx context.Context, contaID string, papel domain.Role, hostname string) (*domain.DominioSite, error) {
	if !podeAdministrarDominio(papel) {
		return nil, errors.New("sem permissão para administrar domínio")
	}
	hostname, err := normalizarHostname(hostname)
	if err != nil {
		return nil, err
	}
	if s.hostsReservados[hostname] {
		return nil, errors.New("este domínio é reservado pela plataforma")
	}
	token := make([]byte, 24)
	if _, err := rand.Read(token); err != nil {
		return nil, fmt.Errorf("gerar token de domínio: %w", err)
	}
	dominio := &domain.DominioSite{ContaID: contaID, Hostname: hostname, TokenVerificacao: hex.EncodeToString(token), Status: "PENDENTE"}
	if err := s.repo.SalvarPendente(ctx, dominio); err != nil {
		return nil, err
	}
	prepararInstrucaoDNS(dominio)
	return dominio, nil
}

func (s *dominioSiteService) Verificar(ctx context.Context, contaID string, papel domain.Role) (*domain.DominioSite, error) {
	if !podeAdministrarDominio(papel) {
		return nil, errors.New("sem permissão para verificar domínio")
	}
	dominio, err := s.repo.ObterPorConta(ctx, contaID)
	if err != nil || dominio == nil {
		return nil, errors.New("configure o domínio antes de verificar")
	}
	nome := "_kaptei-verificacao." + dominio.Hostname
	ctxDNS, cancelar := context.WithTimeout(ctx, 8*time.Second)
	defer cancelar()
	registros, err := s.resolvedor.ConsultarTXT(ctxDNS, nome)
	esperado := "kaptei-verificacao=" + dominio.TokenVerificacao
	if err == nil {
		for _, registro := range registros {
			if strings.TrimSpace(registro) == esperado {
				if err := s.repo.Ativar(ctx, dominio.ID, contaID, dominio.TokenVerificacao); err != nil {
					return nil, err
				}
				return s.Obter(ctx, contaID, papel)
			}
		}
	}
	mensagem := "registro TXT de verificação ainda não foi encontrado"
	if err != nil {
		mensagem = "consulta DNS indisponível ou sem registro TXT"
	}
	if erroRegistro := s.repo.RegistrarFalha(ctx, dominio.ID, contaID, dominio.TokenVerificacao, mensagem); erroRegistro != nil {
		return nil, erroRegistro
	}
	prepararInstrucaoDNS(dominio)
	return dominio, errors.New(mensagem)
}

func (s *dominioSiteService) ResolverPublico(ctx context.Context, hostname string) (*domain.SitePublico, error) {
	hostname, err := normalizarHostname(hostname)
	if err != nil {
		return nil, nil
	}
	return s.repo.ObterSitePorHostname(ctx, hostname)
}

func normalizarHostname(valor string) (string, error) {
	valor = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(valor, ".")))
	if valor == "" || len(valor) > 253 || strings.ContainsAny(valor, "/:@") || net.ParseIP(valor) != nil || valor == "localhost" {
		return "", errors.New("domínio inválido")
	}
	rotulos := strings.Split(valor, ".")
	if len(rotulos) < 2 {
		return "", errors.New("informe um domínio completo")
	}
	for _, rotulo := range rotulos {
		if !rotuloDominioValido.MatchString(rotulo) {
			return "", errors.New("domínio inválido")
		}
	}
	return valor, nil
}

func prepararInstrucaoDNS(dominio *domain.DominioSite) {
	dominio.RegistroTXTNome = "_kaptei-verificacao." + dominio.Hostname
	dominio.RegistroTXTValor = "kaptei-verificacao=" + dominio.TokenVerificacao
}

func podeAdministrarDominio(papel domain.Role) bool {
	return papel == domain.RoleGestor || papel == domain.RoleCorretorSolo || papel == domain.RoleSuperAdmin
}
