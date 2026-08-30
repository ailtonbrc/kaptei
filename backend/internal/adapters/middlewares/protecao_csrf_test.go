package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidarOrigemCookieBloqueiaMutacaoSemOrigem(t *testing.T) {
	proximoChamado := false
	handler := ValidarOrigemCookie([]string{"https://app.kaptei.test"})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		proximoChamado = true
	}))
	requisicao := httptest.NewRequest(http.MethodPost, "https://api.kaptei.test/api/v1/conta", nil)
	requisicao.AddCookie(&http.Cookie{Name: CookieSessao, Value: "sessao"})
	resposta := httptest.NewRecorder()
	handler.ServeHTTP(resposta, requisicao)
	if resposta.Code != http.StatusForbidden || proximoChamado {
		t.Fatalf("esperado bloqueio 403; status=%d proximo=%v", resposta.Code, proximoChamado)
	}
	if !strings.Contains(resposta.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("erro deveria usar JSON: %q", resposta.Header().Get("Content-Type"))
	}
}

func TestValidarOrigemCookiePermiteOrigemConhecida(t *testing.T) {
	handler := ValidarOrigemCookie([]string{"https://app.kaptei.test"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	requisicao := httptest.NewRequest(http.MethodPatch, "https://api.kaptei.test/api/v1/conta", nil)
	requisicao.Header.Set("Origin", "https://app.kaptei.test")
	requisicao.AddCookie(&http.Cookie{Name: CookieSessao, Value: "sessao"})
	resposta := httptest.NewRecorder()
	handler.ServeHTTP(resposta, requisicao)
	if resposta.Code != http.StatusNoContent {
		t.Fatalf("origem conhecida deveria ser aceita; status=%d", resposta.Code)
	}
}

func TestValidarOrigemCookiePermiteClienteBearerSemCookie(t *testing.T) {
	handler := ValidarOrigemCookie([]string{"https://app.kaptei.test"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	requisicao := httptest.NewRequest(http.MethodPost, "https://api.kaptei.test/api/v1/conta", nil)
	requisicao.Header.Set("Authorization", "Bearer token")
	resposta := httptest.NewRecorder()
	handler.ServeHTTP(resposta, requisicao)
	if resposta.Code != http.StatusNoContent {
		t.Fatalf("cliente sem cookie deveria ser aceito; status=%d", resposta.Code)
	}
}
