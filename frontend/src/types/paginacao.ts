export interface ListaPaginada<T> {
  dados: T[];
  total: number;
  pagina: number;
  limite: number;
}

export interface FiltroPaginacao {
  pagina?: number;
  limite?: number;
  busca?: string;
  status?: string;
  tipo?: string;
  finalidade?: string;
}
