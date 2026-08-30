package gateways

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

const limiteRespostaMeta = 512 * 1024

type clienteMetaGraph struct {
	cliente           *http.Client
	baseURL           string
	versao            string
	segredoAplicativo string
}

type erroMetaGraph struct {
	mensagem   string
	permanente bool
}

func (e *erroMetaGraph) Error() string       { return e.mensagem }
func (e *erroMetaGraph) NaoRetentavel() bool { return e.permanente }

func NewClienteMetaGraph(baseURL, versao, segredoAplicativo string, timeout time.Duration) (ports.ClienteMetaGraph, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	versao = strings.Trim(strings.TrimSpace(versao), "/")
	segredoAplicativo = strings.TrimSpace(segredoAplicativo)
	endereco, err := url.Parse(baseURL)
	if err != nil || endereco.Scheme != "https" || endereco.Host == "" || endereco.User != nil || endereco.RawQuery != "" || endereco.Fragment != "" {
		return nil, errors.New("META_GRAPH_BASE_URL deve ser uma URL HTTPS absoluta")
	}
	if !versaoMetaValida(versao) {
		return nil, errors.New("META_GRAPH_API_VERSION deve seguir o formato vNN.N")
	}
	if segredoAplicativo == "" {
		return nil, errors.New("segredo do aplicativo Meta Ã© obrigatÃ³rio")
	}
	if timeout <= 0 || timeout > time.Minute {
		return nil, errors.New("timeout do cliente Meta deve estar entre zero e um minuto")
	}
	cliente := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &clienteMetaGraph{cliente: cliente, baseURL: baseURL, versao: versao, segredoAplicativo: segredoAplicativo}, nil
}

func (c *clienteMetaGraph) ObterLead(ctx context.Context, leadID, tokenPagina string) (*domain.LeadMeta, error) {
	if !somenteDigitos(leadID) || strings.TrimSpace(tokenPagina) == "" {
		return nil, &erroMetaGraph{mensagem: "identificador ou token Meta invÃ¡lido", permanente: true}
	}
	endereco := fmt.Sprintf("%s/%s/%s", c.baseURL, c.versao, url.PathEscape(leadID))
	parametros := url.Values{}
	parametros.Set("fields", "id,created_time,ad_id,form_id,field_data")
	parametros.Set("appsecret_proof", gerarProvaSegredo(tokenPagina, c.segredoAplicativo))
	requisicao, err := http.NewRequestWithContext(ctx, http.MethodGet, endereco+"?"+parametros.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisiÃ§Ã£o Meta Graph: %w", err)
	}
	requisicao.Header.Set("Authorization", "Bearer "+tokenPagina)
	requisicao.Header.Set("Accept", "application/json")
	resposta, err := c.cliente.Do(requisicao)
	if err != nil {
		return nil, fmt.Errorf("consultar Meta Graph: %w", err)
	}
	defer resposta.Body.Close()
	corpo, err := io.ReadAll(io.LimitReader(resposta.Body, limiteRespostaMeta+1))
	if err != nil {
		return nil, fmt.Errorf("ler resposta Meta Graph: %w", err)
	}
	if len(corpo) > limiteRespostaMeta {
		return nil, &erroMetaGraph{mensagem: "resposta Meta Graph excedeu o limite", permanente: true}
	}
	if resposta.StatusCode < 200 || resposta.StatusCode >= 300 {
		mensagem := mensagemErroMeta(corpo, resposta.StatusCode)
		permanente := resposta.StatusCode >= 400 && resposta.StatusCode < 500 && resposta.StatusCode != http.StatusTooManyRequests
		return nil, &erroMetaGraph{mensagem: mensagem, permanente: permanente}
	}
	var dados struct {
		ID           string  `json:"id"`
		CriadoEm     string  `json:"created_time"`
		AnuncioID    *string `json:"ad_id"`
		FormularioID *string `json:"form_id"`
		Campos       []struct {
			Nome    string   `json:"name"`
			Valores []string `json:"values"`
		} `json:"field_data"`
	}
	if err := json.Unmarshal(corpo, &dados); err != nil {
		return nil, &erroMetaGraph{mensagem: "resposta Meta Graph invÃ¡lida", permanente: true}
	}
	if dados.ID == "" {
		return nil, &erroMetaGraph{mensagem: "Meta Graph retornou lead sem identificador", permanente: true}
	}
	lead := &domain.LeadMeta{ID: dados.ID, AnuncioID: dados.AnuncioID, FormularioID: dados.FormularioID, Campos: make(map[string][]string, len(dados.Campos))}
	if instante, ok := interpretarDataMeta(dados.CriadoEm); ok {
		instante = instante.UTC()
		lead.CriadoEm = &instante
	}
	for _, campo := range dados.Campos {
		nome := strings.ToLower(strings.TrimSpace(campo.Nome))
		if nome != "" {
			lead.Campos[nome] = append([]string(nil), campo.Valores...)
		}
	}
	return lead, nil
}

func interpretarDataMeta(valor string) (time.Time, bool) {
	for _, formato := range []string{time.RFC3339, "2006-01-02T15:04:05-0700"} {
		if instante, err := time.Parse(formato, valor); err == nil {
			return instante, true
		}
	}
	return time.Time{}, false
}

func gerarProvaSegredo(token, segredo string) string {
	mac := hmac.New(sha256.New, []byte(segredo))
	_, _ = mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func mensagemErroMeta(corpo []byte, status int) string {
	var resposta struct {
		Erro struct {
			Mensagem string `json:"message"`
			Tipo     string `json:"type"`
			Codigo   int    `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(corpo, &resposta) == nil && resposta.Erro.Mensagem != "" {
		return fmt.Sprintf("Meta Graph recusou a consulta (status %d, cÃ³digo %d, tipo %s): %s", status, resposta.Erro.Codigo, resposta.Erro.Tipo, resposta.Erro.Mensagem)
	}
	return fmt.Sprintf("Meta Graph recusou a consulta com status %d", status)
}

func versaoMetaValida(valor string) bool {
	if len(valor) < 4 || valor[0] != 'v' {
		return false
	}
	partes := strings.Split(valor[1:], ".")
	return len(partes) == 2 && somenteDigitos(partes[0]) && somenteDigitos(partes[1])
}

func somenteDigitos(valor string) bool {
	if valor == "" {
		return false
	}
	for _, caractere := range valor {
		if caractere < '0' || caractere > '9' {
			return false
		}
	}
	return true
}
