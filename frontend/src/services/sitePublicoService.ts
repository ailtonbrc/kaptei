import { api } from './api';
import axios from 'axios';
import type { CapturaLeadSite, EventoConversaoSite, FiltrosCatalogo, ImovelPublico, SitePublico } from '../types/sitePublico';

interface PaginaImoveisPublicos {
  dados: ImovelPublico[];
  total: number;
  pagina: number;
  limite: number;
}

const montarParametros = (filtros: FiltrosCatalogo): URLSearchParams => {
  const parametros = new URLSearchParams();
  Object.entries(filtros).forEach(([chave, valor]) => {
    if (valor !== undefined && valor !== '') parametros.set(chave, String(valor));
  });
  return parametros;
};

export const sitePublicoService = {
  async obter(slug: string): Promise<SitePublico> {
    const resposta = await api.get<SitePublico>(`/public/sites/${encodeURIComponent(slug)}`);
    return resposta.data;
  },
  async obterPorDominio(hostname: string): Promise<SitePublico | undefined> {
    try {
      const resposta = await api.get<SitePublico>(`/public/dominios/${encodeURIComponent(hostname)}`);
      return resposta.data;
    } catch (erro) {
      if (axios.isAxiosError(erro) && erro.response?.status === 404) return undefined;
      throw erro;
    }
  },


  async listarImoveis(slug: string, filtros: FiltrosCatalogo = {}): Promise<PaginaImoveisPublicos> {
    const parametros = montarParametros(filtros);
    const resposta = await api.get<PaginaImoveisPublicos>(
      `/public/sites/${encodeURIComponent(slug)}/imoveis?${parametros.toString()}`,
    );
    return resposta.data;
  },

  async obterImovel(slugSite: string, slugImovel: string): Promise<ImovelPublico> {
    const resposta = await api.get<ImovelPublico>(
      `/public/sites/${encodeURIComponent(slugSite)}/imoveis/${encodeURIComponent(slugImovel)}`,
    );
    return resposta.data;
  },

  async registrarEventoConversao(slug: string, evento: EventoConversaoSite): Promise<void> {
    await api.post(`/public/sites/${encodeURIComponent(slug)}/eventos-conversao`, evento);
  },

  async captarLead(slug: string, dados: CapturaLeadSite): Promise<void> {
    await api.post(`/public/sites/${encodeURIComponent(slug)}/leads`, dados);
  },
};
