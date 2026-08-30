package main

import (
	"html/template"
	"net"
	"strconv"
	"strings"
)

func partesRotaDominio(caminho string) *rotaPublica {
	partes := strings.Split(strings.Trim(caminho, "/"), "/")
	if len(partes) == 1 && partes[0] == "" {
		return &rotaPublica{}
	}
	if len(partes) == 1 && partes[0] == "privacidade" {
		return &rotaPublica{privacidade: true}
	}
	if len(partes) == 2 && partes[0] == "imoveis" && partes[1] != "" {
		return &rotaPublica{imovel: partes[1]}
	}
	return nil
}

func hostnameRequisicao(host string) (string, bool) {
	host = strings.TrimSpace(strings.ToLower(host))
	if strings.Contains(host, ":") {
		hostname, portaTexto, err := net.SplitHostPort(host)
		if err != nil {
			return "", false
		}
		porta, err := strconv.Atoi(portaTexto)
		if err != nil || porta < 1 || porta > 65535 {
			return "", false
		}
		host = hostname
	}
	host = strings.TrimSuffix(host, ".")
	if host == "" || len(host) > 253 || strings.ContainsAny(host, "/:@") {
		return "", false
	}
	for _, caractere := range host {
		if (caractere < 'a' || caractere > 'z') && (caractere < '0' || caractere > '9') && caractere != '.' && caractere != '-' {
			return "", false
		}
	}
	return host, true
}

func ajustarLinksDominio(corpo template.HTML, slug string, dominioPersonalizado bool) template.HTML {
	if !dominioPersonalizado {
		return corpo
	}
	html := string(corpo)
	prefixo := "/s/" + slug
	html = strings.ReplaceAll(html, prefixo+"/imoveis/", "/imoveis/")
	html = strings.ReplaceAll(html, `href="`+prefixo+`"`, `href="/"`)
	return template.HTML(html)
}
