package observabilidade

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/msdev/kaptei/internal/core/domain"
)

type repositorioConfiguracaoTeste struct {
	configuracao *domain.ConfiguracaoSistema
}

func (r repositorioConfiguracaoTeste) Get(context.Context, string) (*domain.ConfiguracaoSistema, error) {
	return r.configuracao, nil
}

func (repositorioConfiguracaoTeste) Set(context.Context, *domain.ConfiguracaoSistema) error {
	return nil
}

type protetorMetricasTeste struct{}

func (protetorMetricasTeste) Proteger(valor string) (string, error) { return "protegido:" + valor, nil }
func (protetorMetricasTeste) Revelar(valor string) (string, error) {
	return valor[len("protegido:"):], nil
}

func TestHandlerMetricasOcultaEndpointDesativado(t *testing.T) {
	valor, _ := json.Marshal(configuracaoAcesso{Ativa: false})
	handler := &handlerProtegido{
		proximo:       http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		configuracoes: repositorioConfiguracaoTeste{configuracao: &domain.ConfiguracaoSistema{Valor: valor}},
		protetor:      protetorMetricasTeste{},
	}
	resposta := httptest.NewRecorder()
	handler.ServeHTTP(resposta, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if resposta.Code != http.StatusNotFound {
		t.Fatalf("endpoint desativado retornou %d", resposta.Code)
	}
}

func TestHandlerMetricasExigeTokenCorreto(t *testing.T) {
	valor, _ := json.Marshal(configuracaoAcesso{Ativa: true, Token: "protegido:token-com-mais-de-trinta-e-dois-caracteres"})
	handler := &handlerProtegido{
		proximo:       http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
		configuracoes: repositorioConfiguracaoTeste{configuracao: &domain.ConfiguracaoSistema{Valor: valor}},
		protetor:      protetorMetricasTeste{},
	}
	semToken := httptest.NewRecorder()
	handler.ServeHTTP(semToken, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if semToken.Code != http.StatusUnauthorized {
		t.Fatalf("requisição sem token retornou %d", semToken.Code)
	}

	requisicao := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	requisicao.Header.Set("Authorization", "Bearer token-com-mais-de-trinta-e-dois-caracteres")
	comToken := httptest.NewRecorder()
	handler.ServeHTTP(comToken, requisicao)
	if comToken.Code != http.StatusOK {
		t.Fatalf("requisição autorizada retornou %d", comToken.Code)
	}
}
