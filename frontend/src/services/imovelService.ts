import { api } from './api';
import type { Imovel, ImovelFoto } from '../types/imovel';
import type { FiltroPaginacao, ListaPaginada } from '../types/paginacao';

export const imovelService = {
  async listar(filtros: FiltroPaginacao = {}): Promise<ListaPaginada<Imovel>> {
	const response = await api.get('/v1/negocios/imoveis', { params: { pagina: 1, limite: 100, ...filtros } });
    return response.data;
  },

  async buscarPorId(id: string): Promise<Imovel> {
    const response = await api.get(`/v1/negocios/imoveis/${id}`);
    return response.data;
  },

  async criar(imovel: Partial<Imovel>): Promise<Imovel> {
    const response = await api.post('/v1/negocios/imoveis', imovel);
    return response.data;
  },

  async atualizar(id: string, imovel: Partial<Imovel>): Promise<void> {
    await api.put(`/v1/negocios/imoveis/${id}`, imovel);
  },

  async deletar(id: string): Promise<void> {
    await api.delete(`/v1/negocios/imoveis/${id}`);
  },

  async adicionarFoto(imovelId: string, url: string, isCapa: boolean): Promise<ImovelFoto> {
    const response = await api.post(`/v1/negocios/imoveis/${imovelId}/fotos`, { url, is_capa: isCapa });
    return response.data;
	},

  async enviarFoto(imovelId: string, arquivo: File, isCapa: boolean): Promise<ImovelFoto> {
    const dados = new FormData();
    dados.append('arquivo', arquivo);
    dados.append('is_capa', String(isCapa));
    const response = await api.post(`/v1/negocios/imoveis/${imovelId}/fotos/upload`, dados);
    return response.data;
  },

	async excluirFoto(imovelId: string, fotoId: string): Promise<void> {
	  await api.delete(`/v1/negocios/imoveis/${imovelId}/fotos/${fotoId}`);
	}
};
