import { api } from './api';
import type { Cliente, Interacao } from '../types/cliente';
import type { FiltroPaginacao, ListaPaginada } from '../types/paginacao';

export const clientesService = {
	listar: async (filtros: FiltroPaginacao = {}): Promise<ListaPaginada<Cliente>> => {
	const response = await api.get('/v1/negocios/clientes', { params: { pagina: 1, limite: 100, ...filtros } });
    return response.data;
  },

  obterPorId: async (id: string): Promise<Cliente> => {
    const response = await api.get(`/v1/negocios/clientes/${id}`);
    return response.data;
  },

  criar: async (cliente: Cliente): Promise<Cliente> => {
    const response = await api.post('/v1/negocios/clientes', cliente);
    return response.data;
  },

  atualizar: async (id: string, cliente: Partial<Cliente>): Promise<Cliente> => {
    const response = await api.put(`/v1/negocios/clientes/${id}`, cliente);
    return response.data;
  },

  excluir: async (id: string): Promise<void> => {
    await api.delete(`/v1/negocios/clientes/${id}`);
  },

  // --- INTERAÇÕES (TIMELINE) ---
  listarInteracoes: async (clienteId: string): Promise<Interacao[]> => {
    const response = await api.get(`/v1/negocios/clientes/${clienteId}/interacoes`);
    return response.data;
  },

  adicionarInteracao: async (clienteId: string, interacao: Partial<Interacao>): Promise<Interacao> => {
    const response = await api.post(`/v1/negocios/clientes/${clienteId}/interacoes`, interacao);
    return response.data;
  },
  
  excluirInteracao: async (id: string): Promise<void> => {
    await api.delete(`/v1/negocios/interacoes/${id}`);
  }
};
