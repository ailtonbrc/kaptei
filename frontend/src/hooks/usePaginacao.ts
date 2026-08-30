import { useState, useCallback, useEffect, useRef } from 'react';

export interface PaginatedResult<T> {
  dados: T[];
  total: number;
  pagina: number;
}

export interface UsePaginacaoOptions<T> {
  fetchData: (pagina: number, busca: string, filtroExtra?: string) => Promise<PaginatedResult<T>>;
  initialSearchTerm?: string;
  initialFilter?: string;
  autoLoad?: boolean;
}

export function usePaginacao<T>({ fetchData, initialSearchTerm = '', initialFilter = '', autoLoad = true }: UsePaginacaoOptions<T>) {
  const [items, setItems] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [pagina, setPagina] = useState(1);
  const [total, setTotal] = useState(0);
  const [searchTerm, setSearchTerm] = useState(initialSearchTerm);
  const [filter, setFilter] = useState(initialFilter);

  const fetchRef = useRef(fetchData);
  useEffect(() => {
    fetchRef.current = fetchData;
  }, [fetchData]);

  const carregarItems = useCallback(async (paginaSolicitada = 1, acumular = false, buscaAtual = searchTerm, filtroAtual = filter) => {
    try {
      setLoading(true);
      const resultado = await fetchRef.current(paginaSolicitada, buscaAtual, filtroAtual);
      setItems((atuais) => acumular ? [...atuais, ...resultado.dados] : resultado.dados);
      setPagina(resultado.pagina);
      setTotal(resultado.total);
    } catch (error) {
      console.error('Erro ao carregar dados:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [searchTerm, filter]);

  useEffect(() => {
    if (autoLoad) {
      void carregarItems(1, false, searchTerm, filter);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [carregarItems, autoLoad]);

  const buscar = () => void carregarItems(1, false, searchTerm, filter);

  const carregarMais = () => {
    if (items.length < total) {
      void carregarItems(pagina + 1, true, searchTerm, filter);
    }
  };

  const removeItem = useCallback((id: string | number) => {
    setItems((atuais) => atuais.filter((item: any) => item.id !== id));
    setTotal((t) => t - 1);
  }, []);

  return {
    items, setItems, loading, pagina, total, searchTerm, setSearchTerm, filter, setFilter, buscar, carregarMais, removeItem, carregarItems
  };
}
