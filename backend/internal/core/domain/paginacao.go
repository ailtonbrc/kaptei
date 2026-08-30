package domain

type FiltroPaginacao struct {
	Pagina     int
	Limite     int
	Busca      string
	Status     string
	Tipo       string
	Finalidade string
	UsuarioID  *string
}

type ListaPaginada[T any] struct {
	Dados  []T `json:"dados"`
	Total  int `json:"total"`
	Pagina int `json:"pagina"`
	Limite int `json:"limite"`
}
