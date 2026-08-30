package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/msdev/kaptei/internal/core/domain"
)

func filtroPaginacao(r *http.Request) domain.FiltroPaginacao {
	pagina := inteiroConsulta(r, "pagina", 1)
	limite := inteiroConsulta(r, "limite", 50)
	if pagina < 1 {
		pagina = 1
	}
	if limite < 1 {
		limite = 50
	}
	if limite > 100 {
		limite = 100
	}
	return domain.FiltroPaginacao{
		Pagina: pagina, Limite: limite,
		Busca:      strings.TrimSpace(r.URL.Query().Get("busca")),
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		Tipo:       strings.TrimSpace(r.URL.Query().Get("tipo")),
		Finalidade: strings.TrimSpace(r.URL.Query().Get("finalidade")),
	}
}

func inteiroConsulta(r *http.Request, chave string, padrao int) int {
	valor, err := strconv.Atoi(r.URL.Query().Get(chave))
	if err != nil {
		return padrao
	}
	return valor
}
