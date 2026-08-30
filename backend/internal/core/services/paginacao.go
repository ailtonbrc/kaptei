package services

import (
	"strings"

	"github.com/msdev/kaptei/internal/core/domain"
)

func normalizarFiltroPaginacao(filtro domain.FiltroPaginacao) domain.FiltroPaginacao {
	if filtro.Pagina < 1 {
		filtro.Pagina = 1
	}
	if filtro.Limite < 1 {
		filtro.Limite = 50
	}
	if filtro.Limite > 100 {
		filtro.Limite = 100
	}
	filtro.Busca = limitarRunas(strings.TrimSpace(filtro.Busca), 120)
	filtro.Status = limitarRunas(strings.TrimSpace(filtro.Status), 40)
	filtro.Tipo = limitarRunas(strings.TrimSpace(filtro.Tipo), 40)
	filtro.Finalidade = limitarRunas(strings.TrimSpace(filtro.Finalidade), 40)
	return filtro
}

func limitarRunas(valor string, limite int) string {
	valor = strings.TrimSpace(valor)
	runas := []rune(valor)
	if len(runas) > limite {
		return string(runas[:limite])
	}
	return valor
}
