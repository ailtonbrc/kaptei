import { api } from './api';
import type { Lead } from '../types/lead';
import type { FiltroPaginacao, ListaPaginada } from '../types/paginacao';

export const leadsService = {
  listar: async (filtros: FiltroPaginacao = {}): Promise<ListaPaginada<Lead>> => {
    try {
	  const response = await api.get('/v1/negocios/leads', { params: { pagina: 1, limite: 100, ...filtros } });
      return response.data;
    } catch (error) {
      console.error('Erro ao listar leads:', error);
      throw error;
    }
  },

  atribuir: async (id: string, usuario_id: string): Promise<void> => {
    try {
      await api.post(`/v1/negocios/leads/${id}/atribuir`, { usuario_id });
    } catch (error) {
      console.error('Erro ao atribuir lead:', error);
      throw error;
    }
  },

  qualificar: async (id: string): Promise<void> => {
    try {
      await api.post(`/v1/negocios/leads/${id}/qualificar`);
    } catch (error) {
      console.error('Erro ao qualificar lead:', error);
      throw error;
    }
  },

  descartar: async (id: string, motivo: string): Promise<void> => {
    try {
      await api.post(`/v1/negocios/leads/${id}/descartar`, { motivo });
    } catch (error) {
      console.error('Erro ao descartar lead:', error);
      throw error;
    }
  }
};
