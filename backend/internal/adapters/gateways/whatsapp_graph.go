package gateways

import (
	"bytes"
	"context"
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

type clienteWhatsAppGraph struct {
	cliente           *http.Client
	baseURL           string
	versao            string
	segredoAplicativo string
}

func NewClienteWhatsAppGraph(baseURL, versao, segredoAplicativo string, timeout time.Duration) (ports.ClienteWhatsApp, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	versao = strings.Trim(strings.TrimSpace(versao), "/")
	segredoAplicativo = strings.TrimSpace(segredoAplicativo)
	endereco, err := url.Parse(baseURL)
	if err != nil || endereco.Scheme != "https" || endereco.Host == "" || endereco.User != nil || endereco.RawQuery != "" || endereco.Fragment != "" {
		return nil, errors.New("URL base do WhatsApp Graph inválida")
	}
	if !versaoMetaValida(versao) || segredoAplicativo == "" || timeout <= 0 || timeout > time.Minute {
		return nil, errors.New("configuração do cliente WhatsApp Graph inválida")
	}
	cliente := &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return &clienteWhatsAppGraph{cliente: cliente, baseURL: baseURL, versao: versao, segredoAplicativo: segredoAplicativo}, nil
}

func (c *clienteWhatsAppGraph) Enviar(ctx context.Context, numeroTelefoneID, tokenAcesso string, solicitacao *domain.SolicitacaoEnvioWhatsApp) (string, error) {
	if !somenteDigitos(numeroTelefoneID) || strings.TrimSpace(tokenAcesso) == "" || solicitacao == nil {
		return "", &erroMetaGraph{mensagem: "solicitação WhatsApp inválida", permanente: true}
	}
	payload := map[string]any{"messaging_product": "whatsapp", "recipient_type": "individual", "to": solicitacao.Destinatario}
	switch solicitacao.Tipo {
	case "TEXTO":
		payload["type"] = "text"
		payload["text"] = map[string]any{"preview_url": false, "body": solicitacao.Texto}
	case "TEMPLATE":
		payload["type"] = "template"
		template := map[string]any{"name": solicitacao.TemplateNome, "language": map[string]string{"code": solicitacao.TemplateIdioma}}
		if len(solicitacao.Parametros) > 0 {
			parametros := make([]map[string]string, 0, len(solicitacao.Parametros))
			for _, valor := range solicitacao.Parametros {
				parametros = append(parametros, map[string]string{"type": "text", "text": valor})
			}
			template["components"] = []any{map[string]any{"type": "body", "parameters": parametros}}
		}
		payload["template"] = template
	default:
		return "", &erroMetaGraph{mensagem: "tipo de mensagem WhatsApp não suportado", permanente: true}
	}
	corpo, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	endereco := fmt.Sprintf("%s/%s/%s/messages?appsecret_proof=%s", c.baseURL, c.versao, url.PathEscape(numeroTelefoneID), gerarProvaSegredo(tokenAcesso, c.segredoAplicativo))
	requisicao, err := http.NewRequestWithContext(ctx, http.MethodPost, endereco, bytes.NewReader(corpo))
	if err != nil {
		return "", fmt.Errorf("criar requisição WhatsApp: %w", err)
	}
	requisicao.Header.Set("Authorization", "Bearer "+tokenAcesso)
	requisicao.Header.Set("Content-Type", "application/json")
	requisicao.Header.Set("Accept", "application/json")
	resposta, err := c.cliente.Do(requisicao)
	if err != nil {
		return "", fmt.Errorf("enviar mensagem WhatsApp: %w", err)
	}
	defer resposta.Body.Close()
	respostaCorpo, err := io.ReadAll(io.LimitReader(resposta.Body, limiteRespostaMeta+1))
	if err != nil {
		return "", fmt.Errorf("ler resposta WhatsApp: %w", err)
	}
	if len(respostaCorpo) > limiteRespostaMeta {
		return "", &erroMetaGraph{mensagem: "resposta WhatsApp excedeu o limite", permanente: true}
	}
	if resposta.StatusCode < 200 || resposta.StatusCode >= 300 {
		permanente := resposta.StatusCode >= 400 && resposta.StatusCode < 500 && resposta.StatusCode != http.StatusTooManyRequests
		return "", &erroMetaGraph{mensagem: mensagemErroMeta(respostaCorpo, resposta.StatusCode), permanente: permanente}
	}
	var dados struct {
		Mensagens []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if json.Unmarshal(respostaCorpo, &dados) != nil || len(dados.Mensagens) == 0 || dados.Mensagens[0].ID == "" || len(dados.Mensagens[0].ID) > 255 {
		return "", &erroMetaGraph{mensagem: "resposta de envio WhatsApp inválida", permanente: true}
	}
	return dados.Mensagens[0].ID, nil
}
