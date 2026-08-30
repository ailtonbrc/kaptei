package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
)

type repositorioRetencaoFalso struct {
	politica      *domain.PoliticaRetencao
	politicaSalva *domain.PoliticaRetencao
	executou      bool
	bloqueioSalvo *domain.BloqueioRetencao
	erro          error
}

func (r *repositorioRetencaoFalso) ObterPolitica(context.Context, string) (*domain.PoliticaRetencao, error) {
	if r.erro != nil {
		return nil, r.erro
	}
	if r.politica == nil {
		return &domain.PoliticaRetencao{}, nil
	}
	copia := *r.politica
	return &copia, nil
}

func (r *repositorioRetencaoFalso) SalvarPolitica(_ context.Context, politica *domain.PoliticaRetencao, _ string) error {
	copia := *politica
	r.politicaSalva = &copia
	return r.erro
}

func (r *repositorioRetencaoFalso) GerarRelatorio(context.Context, string, *domain.PoliticaRetencao) (*domain.RelatorioRetencao, error) {
	return &domain.RelatorioRetencao{LeadsElegiveis: 2}, r.erro
}

func (r *repositorioRetencaoFalso) Executar(context.Context, string, string, *domain.PoliticaRetencao) (*domain.ResultadoRetencao, error) {
	r.executou = true
	return &domain.ResultadoRetencao{LeadsAnonimizados: 1}, r.erro
}

func (r *repositorioRetencaoFalso) ListarBloqueios(context.Context, string) ([]domain.BloqueioRetencao, error) {
	return []domain.BloqueioRetencao{}, r.erro
}

func (r *repositorioRetencaoFalso) SalvarBloqueio(_ context.Context, bloqueio *domain.BloqueioRetencao, _ string) error {
	bloqueio.ID = "00000000-0000-4000-8000-000000000001"
	bloqueio.CriadoEm = time.Now()
	copia := *bloqueio
	r.bloqueioSalvo = &copia
	return r.erro
}

func (r *repositorioRetencaoFalso) RemoverBloqueio(context.Context, string, string) error {
	return r.erro
}

func politicaRetencaoValida() domain.PoliticaRetencao {
	return domain.PoliticaRetencao{
		Ativa:                true,
		DiasLeadsDescartados: 730,
		DiasClientesPerdidos: 1825,
		TamanhoLote:          200,
		FundamentoLegal:      "Término da finalidade e ausência de obrigação de conservação.",
	}
}

func TestRetencaoRejeitaPapelSemPermissao(t *testing.T) {
	repositorio := &repositorioRetencaoFalso{politica: &domain.PoliticaRetencao{}}
	servico := NewRetencaoService(repositorio)

	if _, err := servico.ObterPolitica(context.Background(), "conta-1", domain.RoleCorretorEquipe); err == nil {
		t.Fatal("era esperado erro de permissão")
	}
}

func TestRetencaoValidaLimitesAntesDeSalvar(t *testing.T) {
	repositorio := &repositorioRetencaoFalso{}
	servico := NewRetencaoService(repositorio)
	politica := politicaRetencaoValida()
	politica.DiasLeadsDescartados = 29

	err := servico.SalvarPolitica(context.Background(), "conta-1", "usuario-1", domain.RoleGestor, politica)
	if err == nil {
		t.Fatal("era esperado erro para prazo inferior ao mínimo")
	}
	if repositorio.politicaSalva != nil {
		t.Fatal("política inválida não deve chegar ao repositório")
	}
}

func TestRetencaoExigeConfirmacaoExataEPoliticaAtiva(t *testing.T) {
	politica := politicaRetencaoValida()
	repositorio := &repositorioRetencaoFalso{politica: &politica}
	servico := NewRetencaoService(repositorio)

	if _, err := servico.Executar(context.Background(), "conta-1", "usuario-1", domain.RoleGestor, "confirmar"); err == nil {
		t.Fatal("era esperado erro para confirmação inválida")
	}
	if repositorio.executou {
		t.Fatal("repositório não deve executar sem confirmação exata")
	}

	politica.Ativa = false
	if _, err := servico.Executar(context.Background(), "conta-1", "usuario-1", domain.RoleGestor, confirmacaoExecucaoRetencao); err == nil {
		t.Fatal("era esperado erro para política desativada")
	}
}

func TestRetencaoExecutaComConfirmacaoEPoliticaAtiva(t *testing.T) {
	politica := politicaRetencaoValida()
	repositorio := &repositorioRetencaoFalso{politica: &politica}
	servico := NewRetencaoService(repositorio)

	resultado, err := servico.Executar(context.Background(), "conta-1", "usuario-1", domain.RoleGestor, confirmacaoExecucaoRetencao)
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if !repositorio.executou || resultado.LeadsAnonimizados != 1 {
		t.Fatal("execução válida não foi encaminhada")
	}
}

func TestRetencaoValidaBloqueioLegal(t *testing.T) {
	repositorio := &repositorioRetencaoFalso{}
	servico := NewRetencaoService(repositorio)
	bloqueio := domain.BloqueioRetencao{TipoRecurso: "cliente", RecursoID: "00000000-0000-4000-8000-000000000001", Motivo: "Litígio em andamento"}

	salvo, err := servico.SalvarBloqueio(context.Background(), "conta-1", "usuario-1", domain.RoleGestor, bloqueio)
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if salvo.TipoRecurso != "CLIENTE" || repositorio.bloqueioSalvo == nil {
		t.Fatal("bloqueio não foi normalizado e persistido")
	}
}

func TestRetencaoPropagaErroDoRepositorio(t *testing.T) {
	erroEsperado := errors.New("falha de banco")
	repositorio := &repositorioRetencaoFalso{erro: erroEsperado}
	servico := NewRetencaoService(repositorio)

	_, err := servico.GerarRelatorio(context.Background(), "conta-1", domain.RoleGestor)
	if !errors.Is(err, erroEsperado) {
		t.Fatalf("erro inesperado: %v", err)
	}
}
